/// <reference types="@sveltejs/kit" />
/// <reference lib="webworker" />

import { build, files, version } from '$service-worker';

const sw = self as unknown as ServiceWorkerGlobalScope;

const CACHE_NAME = `reminder-${version}`;
const ASSETS = [
	...new Set([
		...build,
		...files,
		'/',
		'/events',
		'/icons/favicon-32.png',
		'/icons/icon-192.png',
		'/icons/icon-512.png',
		'/icons/apple-touch-icon.png',
		'/site.webmanifest'
	])
];

sw.addEventListener('install', (event) => {
	event.waitUntil(
		(async () => {
			try {
				const cache = await caches.open(CACHE_NAME);
				const results = await Promise.allSettled(ASSETS.map((asset) => cache.add(asset)));
				const failed = results.filter((result) => result.status === 'rejected');
				if (failed.length > 0) {
					console.warn(
						`[service-worker] skipped ${failed.length} asset(s) during precache`,
						failed
					);
				}
			} catch (error) {
				console.warn('[service-worker] precache skipped due to storage/cache error', error);
			}
		})()
	);
	sw.skipWaiting();
});

sw.addEventListener('activate', (event) => {
	event.waitUntil(
		caches.keys().then(async (keys) => {
			await Promise.all(keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key)));
			await sw.clients.claim();
		})
	);
});

sw.addEventListener('message', (event) => {
	if (event.data?.type === 'SKIP_WAITING') {
		void sw.skipWaiting();
	}
});

sw.addEventListener('fetch', (event) => {
	if (event.request.method !== 'GET') {
		return;
	}

	const request = event.request;
	const url = new URL(request.url);
	const isSameOrigin = url.origin === sw.location.origin;
	const isPageRequest = request.mode === 'navigate';
	const isBackendRequest =
		isSameOrigin &&
		(url.pathname.startsWith('/api/') ||
			url.pathname === '/healthz' ||
			url.pathname === '/openapi.yaml' ||
			url.pathname.startsWith('/ota/'));
	const isStaticAssetRequest =
		isSameOrigin &&
		(request.destination === 'style' ||
			request.destination === 'script' ||
			request.destination === 'font' ||
			request.destination === 'image' ||
			request.destination === 'manifest' ||
			build.includes(url.pathname) ||
			files.includes(url.pathname));

	if (isPageRequest) {
		event.respondWith(
			fetch(request).catch(async () => {
				return (
					(await caches.match('/')) ??
					(await caches.match('/events')) ??
					new Response('Offline', {
						status: 503,
						headers: { 'Content-Type': 'text/plain; charset=utf-8' }
					})
				);
			})
		);
		return;
	}

	if (isBackendRequest || !isStaticAssetRequest) {
		event.respondWith(fetch(request));
		return;
	}

	event.respondWith(
		caches.match(request).then((cached) => {
			if (cached) {
				return cached;
			}

			return fetch(request).then((response) => {
				if (!response.ok || response.type === 'opaque') {
					return response;
				}

				const copy = response.clone();
				void caches.open(CACHE_NAME).then((cache) => {
					void cache.put(request, copy);
				});
				return response;
			});
		})
	);
});
