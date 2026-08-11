import type { AdminUser, UserRole } from '$lib/auth/auth-store.svelte';
import { toastStore } from '$lib/ui/toast-store.svelte';
import { createUser, listUsers, updateUser } from './api';

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

	async create(displayName: string, username: string, password: string, role: UserRole) {
		if (this.isSaving) return false;
		this.isSaving = true;
		try {
			const item = await createUser(displayName, username, password, role);
			this.items = sortUsers([...this.items, item]);
			this.loaded = true;
			this.lastLoadedAt = Date.now();
			toastStore.success('Đã tạo tài khoản');
			return true;
		} catch (error) {
			toastStore.error(message(error));
			return false;
		} finally {
			this.isSaving = false;
		}
	}

	async update(
		id: string,
		displayName: string,
		username: string,
		role: UserRole,
		password: string
	) {
		if (this.isSaving) return null;
		this.isSaving = true;
		try {
			const item = await updateUser(id, displayName, username, role, password);
			this.items = sortUsers(
				this.items.map((candidate) => (candidate.id === item.id ? item : candidate))
			);
			this.lastLoadedAt = Date.now();
			toastStore.success('Đã cập nhật tài khoản');
			return item;
		} catch (error) {
			toastStore.error(message(error));
			return null;
		} finally {
			this.isSaving = false;
		}
	}

	sync(item: AdminUser) {
		if (!this.loaded) return;
		this.items = this.items.map((candidate) => (candidate.id === item.id ? item : candidate));
	}
}

function sortUsers(items: AdminUser[]) {
	return items.sort((left, right) => left.display_name.localeCompare(right.display_name, 'vi'));
}

function message(error: unknown) {
	return error instanceof Error ? error.message : 'Không thể tải tài khoản';
}

export const userStore = new UserStore();
