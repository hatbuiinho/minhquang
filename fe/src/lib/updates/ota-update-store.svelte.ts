import { Capacitor } from '@capacitor/core';
import { CapacitorUpdater } from '@capgo/capacitor-updater';

import { apiRequest } from '$lib/api/client';

type AppUpdate = {
	available: boolean;
	version?: string;
	url?: string;
};

type OTAStatus = 'idle' | 'checking' | 'downloading' | 'ready' | 'unavailable' | 'error';

class OTAUpdateStore {
	status = $state<OTAStatus>('idle');
	error = $state('');
	currentVersion = $state('');
	nextVersion = $state('');

	async checkOnce() {
		if (!Capacitor.isNativePlatform() || Capacitor.getPlatform() !== 'android') return;
		if (this.status === 'checking' || this.status === 'downloading' || this.status === 'ready')
			return;

		this.status = 'checking';
		this.error = '';

		try {
			await CapacitorUpdater.notifyAppReady();

			const current = await CapacitorUpdater.current();
			this.currentVersion = current.bundle.version;

			const params = new URLSearchParams({
				channel: import.meta.env.VITE_OTA_CHANNEL || 'dev',
				current_version: current.bundle.version,
				native_version: import.meta.env.VITE_NATIVE_VERSION || current.native || '1.0'
			});
			const update = await apiRequest<AppUpdate>(`/api/app-updates/android/latest?${params}`);

			if (!update.available || !update.version || !update.url) {
				this.status = 'unavailable';
				return;
			}

			this.status = 'downloading';
			const bundle = await CapacitorUpdater.download({
				version: update.version,
				url: update.url
			});
			await CapacitorUpdater.next({ id: bundle.id });

			this.nextVersion = update.version;
			this.status = 'ready';
		} catch (err) {
			this.error = err instanceof Error ? err.message : 'Không thể cập nhật OTA';
			this.status = 'error';
			console.error('ota update failed', err);
		}
	}
}

export const otaUpdateStore = new OTAUpdateStore();
