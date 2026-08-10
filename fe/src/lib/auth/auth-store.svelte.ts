import { apiRequest, setAccessToken } from '$lib/api/client';

export type AdminUser = {
	id: string;
	username: string;
	display_name: string;
	role: 'admin';
	active: boolean;
	created_at: string;
	updated_at: string;
};

const tokenKey = 'minhquang_access_token';

class AuthStore {
	user = $state<AdminUser | null>(null);
	initializing = $state(true);
	isSubmitting = $state(false);
	error = $state('');

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
}

export const authStore = new AuthStore();
