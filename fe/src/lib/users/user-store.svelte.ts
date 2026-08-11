import type { AdminUser } from '$lib/auth/auth-store.svelte';
import { toastStore } from '$lib/ui/toast-store.svelte';
import { createUser, listUsers } from './api';

class UserStore {
	items = $state<AdminUser[]>([]);
	isLoading = $state(false);
	isSaving = $state(false);
	loaded = $state(false);
	lastLoadedAt = $state(0);
	private requestGeneration = 0;

	async load() {
		const generation = ++this.requestGeneration;
		this.isLoading = !this.loaded;
		try {
			const items = await listUsers();
			if (generation !== this.requestGeneration) return;
			this.items = items;
			this.loaded = true;
			this.lastLoadedAt = Date.now();
		} catch (error) {
			if (generation === this.requestGeneration) toastStore.error(message(error));
		} finally {
			if (generation === this.requestGeneration) this.isLoading = false;
		}
	}

	async refreshIfStale(ttlMs: number) {
		if (this.loaded && Date.now() - this.lastLoadedAt < ttlMs) return;
		await this.load();
	}

	async create(displayName: string, username: string, password: string) {
		if (this.isSaving) return false;
		this.isSaving = true;
		try {
			await createUser(displayName, username, password);
			toastStore.success('Đã tạo tài khoản quản trị');
			await this.load();
			return true;
		} catch (error) {
			toastStore.error(message(error));
			return false;
		} finally {
			this.isSaving = false;
		}
	}
}

function message(error: unknown) {
	return error instanceof Error ? error.message : 'Không thể tải tài khoản';
}

export const userStore = new UserStore();
