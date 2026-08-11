import { apiRequest } from '$lib/api/client';

export type Department = {
	id: string;
	name: string;
	active: boolean;
	volunteer_count: number;
	active_volunteer_count: number;
	created_at: string;
	updated_at: string;
};

export async function listDepartments(query = '', active: 'all' | 'true' | 'false' = 'all') {
	const params = new URLSearchParams({ active });
	if (query) params.set('q', query);
	const result = await apiRequest<{ departments: Department[] }>(`/api/departments?${params}`);
	return result.departments;
}

export function createDepartment(name: string) {
	return apiRequest<Department>('/api/departments', {
		method: 'POST',
		body: JSON.stringify({ name })
	});
}

export function updateDepartment(id: string, name: string) {
	return apiRequest<Department>(`/api/departments/${encodeURIComponent(id)}`, {
		method: 'PUT',
		body: JSON.stringify({ name })
	});
}

export function setDepartmentActive(id: string, active: boolean) {
	return apiRequest<Department>(`/api/departments/${encodeURIComponent(id)}/status`, {
		method: 'PATCH',
		body: JSON.stringify({ active })
	});
}

export function deleteDepartment(id: string) {
	return apiRequest<void>(`/api/departments/${encodeURIComponent(id)}`, { method: 'DELETE' });
}
