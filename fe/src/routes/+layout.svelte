<script lang="ts">
	import { browser, dev } from '$app/environment';
	import { onMount } from 'svelte';

	import './layout.css';

	let { children } = $props();

	onMount(() => {
		if (!browser || !('serviceWorker' in navigator)) {
			return;
		}

		const capacitor = (window as unknown as { Capacitor?: { isNativePlatform?: () => boolean } })
			.Capacitor;
		const isNativeApp = capacitor?.isNativePlatform?.() === true;
		if (isNativeApp) {
			return;
		}

		if (dev) {
			void navigator.serviceWorker.getRegistrations().then((registrations) => {
				for (const registration of registrations) {
					void registration.unregister();
				}
			});
			return;
		}

		void navigator.serviceWorker.register('/service-worker.js').catch((error) => {
			console.warn('[service-worker] registration failed', error);
		});
	});
</script>

<svelte:head>
	<title>Nhắc việc</title>
	<meta name="description" content="Ứng dụng quản lý sự kiện và nhắc hẹn." />
	<meta name="theme-color" content="#2f6f63" />
	<meta name="mobile-web-app-capable" content="yes" />
	<meta name="apple-mobile-web-app-capable" content="yes" />
	<meta name="apple-mobile-web-app-status-bar-style" content="default" />
	<meta name="apple-mobile-web-app-title" content="Nhắc việc" />
	<meta name="application-name" content="Nhắc việc" />
	<link rel="icon" href="/icons/favicon-32.png" sizes="32x32" type="image/png" />
	<link rel="icon" href="/icons/icon-192.png" sizes="192x192" type="image/png" />
	<link rel="apple-touch-icon" href="/icons/apple-touch-icon.png" />
	<link rel="manifest" href="/site.webmanifest" />
</svelte:head>

<main class="min-h-screen bg-[var(--color-bg)] text-[var(--color-text)]">
	{@render children()}
</main>
