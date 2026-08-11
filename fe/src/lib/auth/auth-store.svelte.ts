import { apiRequest, setAccessToken } from '$lib/api/client';
import { toastStore } from '$lib/ui/toast-store.svelte';

export type Permission =
	| 'volunteer.read'
	| 'volunteer.create'
	| 'volunteer.update'
	| 'volunteer.delete'
	| 'department.read'
	| 'department.manage'
	| 'user.read'
	| 'user.manage';

export type UserRole = 'admin' | 'viewer';

export type AdminUser = {
	id: string;
	username: string;
	display_name: string;
	avatar_url: string;
	role: UserRole;
	permissions: Permission[];
	active: boolean;
	created_at: string;
	updated_at: string;
};

const tokenKey = 'minhquang_access_token';

class AuthStore {
	user = $state<AdminUser | null>(null);
	initializing = $state(true);
	isSubmitting = $state(false);
	isChangingPassword = $state(false);
	isUpdatingProfile = $state(false);
	isUpdatingAvatar = $state(false);
	error = $state('');

	can(permission: Permission): boolean {
		return this.user?.permissions?.includes(permission) ?? false;
	}

	syncUser(item: AdminUser) {
		if (this.user?.id === item.id) this.user = item;
	}

	async init() {
		const token = localStorage.getItem(tokenKey) ?? '';
		if (!token) {
			this.initializing = false;
			return;
		}
		setAccessToken(token);
		try {
			this.user = await apiRequest<AdminUser>('/api/auth/me');
		} catch {
			localStorage.removeItem(tokenKey);
			setAccessToken('');
		} finally {
			this.initializing = false;
		}
	}

	async login(username: string, password: string): Promise<boolean> {
		if (this.isSubmitting) return false;
		this.isSubmitting = true;
		this.error = '';
		try {
			const result = await apiRequest<{ token: string; user: AdminUser }>('/api/auth/login', {
				method: 'POST',
				body: JSON.stringify({ username, password })
			});
			localStorage.setItem(tokenKey, result.token);
			setAccessToken(result.token);
			this.user = result.user;
			return true;
		} catch (error) {
			this.error = error instanceof Error ? error.message : 'Không thể đăng nhập';
			return false;
		} finally {
			this.isSubmitting = false;
		}
	}

	async logout() {
		try {
			await apiRequest<void>('/api/auth/logout', { method: 'POST' });
		} catch {
			// Local logout still proceeds when the server is unavailable.
		}
		localStorage.removeItem(tokenKey);
		setAccessToken('');
		this.user = null;
	}

	async changePassword(currentPassword: string, newPassword: string): Promise<boolean> {
		if (this.isChangingPassword) return false;
		this.isChangingPassword = true;
		try {
			await apiRequest<void>('/api/auth/password', {
				method: 'PATCH',
				body: JSON.stringify({
					current_password: currentPassword,
					new_password: newPassword
				})
			});
			toastStore.success('Đã đổi mật khẩu');
			return true;
		} catch (error) {
			toastStore.error(error instanceof Error ? error.message : 'Không thể đổi mật khẩu');
			return false;
		} finally {
			this.isChangingPassword = false;
		}
	}

	async updateProfile(username: string, displayName: string): Promise<AdminUser | null> {
		if (this.isUpdatingProfile) return null;
		this.isUpdatingProfile = true;
		try {
			const item = await apiRequest<AdminUser>('/api/auth/profile', {
				method: 'PATCH',
				body: JSON.stringify({ username, display_name: displayName })
			});
			this.user = item;
			toastStore.success('Đã cập nhật hồ sơ');
			return item;
		} catch (error) {
			toastStore.error(error instanceof Error ? error.message : 'Không thể cập nhật hồ sơ');
			return null;
		} finally {
			this.isUpdatingProfile = false;
		}
	}

	async updateAvatar(avatarUrl: string): Promise<AdminUser | null> {
		if (this.isUpdatingAvatar) return null;
		this.isUpdatingAvatar = true;
		try {
			const item = await apiRequest<AdminUser>('/api/auth/avatar', {
				method: 'PATCH',
				body: JSON.stringify({ avatar_url: avatarUrl })
			});
			this.user = item;
			toastStore.success('Đã cập nhật ảnh đại diện');
			return item;
		} catch (error) {
			toastStore.error(error instanceof Error ? error.message : 'Không thể cập nhật ảnh đại diện');
			return null;
		} finally {
			this.isUpdatingAvatar = false;
		}
	}
}

export const authStore = new AuthStore();
