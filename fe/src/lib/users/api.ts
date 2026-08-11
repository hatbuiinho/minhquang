import { apiRequest } from '$lib/api/client';
import type { AdminUser } from '$lib/auth/auth-store.svelte';

export async function listUsers() {
	return (await apiRequest<{ users: AdminUser[] }>('/api/users')).users;
}

export function createUser(displayName: string, username: string, password: string) {
	return apiRequest<AdminUser>('/api/users', {
		method: 'POST',
		body: JSON.stringify({ username, display_name: displayName, password })
	});
}
