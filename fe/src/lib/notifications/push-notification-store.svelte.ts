import { Capacitor } from '@capacitor/core';
import { PushNotifications } from '@capacitor/push-notifications';

import { registerDevice, type DevicePlatform } from '$lib/events/api';

type PushStatus = 'idle' | 'unsupported' | 'requesting' | 'registered' | 'denied' | 'error';
type DeviceSyncStatus = 'idle' | 'syncing' | 'synced' | 'error';

class PushNotificationStore {
	status = $state<PushStatus>('idle');
	deviceSyncStatus = $state<DeviceSyncStatus>('idle');
	token = $state('');
	error = $state('');
	syncError = $state('');
	lastMessage = $state('');
	private listenersReady = false;

	async register() {
		if (!Capacitor.isNativePlatform()) {
			this.status = 'unsupported';
			this.error = 'Thông báo đẩy chỉ chạy trên app Android/iOS.';
			return;
		}

		this.status = 'requesting';
		this.error = '';
		this.bindListeners();

		const permission = await PushNotifications.requestPermissions();
		if (permission.receive !== 'granted') {
			this.status = 'denied';
			this.error = 'Bạn chưa cấp quyền nhận thông báo.';
			return;
		}

		await PushNotifications.register();
	}

	private bindListeners() {
		if (this.listenersReady) return;
		this.listenersReady = true;

		void PushNotifications.addListener('registration', (token) => {
			this.token = token.value;
			this.status = 'registered';
			this.error = '';
			void this.syncDeviceToken(token.value);
		});

		void PushNotifications.addListener('registrationError', (error) => {
			this.status = 'error';
			this.error = error.error;
		});

		void PushNotifications.addListener('pushNotificationReceived', (notification) => {
			this.lastMessage = notification.title ?? notification.body ?? 'Đã nhận thông báo mới';
		});

		void PushNotifications.addListener('pushNotificationActionPerformed', (action) => {
			this.lastMessage =
				action.notification.title ?? action.notification.body ?? 'Bạn vừa mở một thông báo';
		});
	}

	private async syncDeviceToken(token: string) {
		const platform = Capacitor.getPlatform();
		if (platform !== 'android' && platform !== 'ios' && platform !== 'web') {
			return;
		}

		this.deviceSyncStatus = 'syncing';
		this.syncError = '';
		try {
			await registerDevice({
				platform: platform as DevicePlatform,
				push_token: token
			});
			this.deviceSyncStatus = 'synced';
		} catch (err) {
			this.deviceSyncStatus = 'error';
			this.syncError = err instanceof Error ? err.message : 'Không đồng bộ được thiết bị';
		}
	}
}

export const pushNotificationStore = new PushNotificationStore();
