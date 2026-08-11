import { apiRequest } from '$lib/api/client';

export type Volunteer = {
	id: string;
	full_name: string;
	dharma_name: string;
	birth_date: string;
	cultivation_place: string;
	phone: string;
	department_id?: string;
	department: string;
	notes: string;
	avatar_url: string;
	arrival_date: string;
	departure_date?: string;
	status: 'active' | 'departed';
	created_at: string;
	updated_at: string;
};

export type VolunteerInput = {
	full_name: string;
	dharma_name: string;
	birth_date: string;
	cultivation_place: string;
	phone: string;
	department: string;
	notes: string;
	avatar_url: string;
	arrival_date: string;
	departure_date: string;
};

export type VolunteerSortKey =
	| 'full_name'
	| 'dharma_name'
	| 'birth_date'
	| 'cultivation_place'
	| 'department'
	| 'phone'
	| 'arrival_date'
	| 'departure_date'
	| 'status';

export type VolunteerPage = {
	volunteers: Volunteer[];
	total: number;
	has_more: boolean;
};

export type VolunteerBulkField =
	| 'full_name'
	| 'dharma_name'
	| 'birth_date'
	| 'cultivation_place'
	| 'phone'
	| 'department'
	| 'notes'
	| 'avatar_url'
	| 'arrival_date'
	| 'departure_date';

export async function listVolunteers(
	query = '',
	status = '',
	departmentId = '',
	limit = 20,
	offset = 0,
	sortBy: VolunteerSortKey = 'arrival_date',
	sortDirection: 'asc' | 'desc' = 'desc'
): Promise<VolunteerPage> {
	const params = new URLSearchParams();
	if (query) params.set('q', query);
	if (status) params.set('status', status);
	if (departmentId) params.set('department_id', departmentId);
	params.set('limit', String(limit));
	params.set('offset', String(offset));
	params.set('sort_by', sortBy);
	params.set('sort_direction', sortDirection);
	return apiRequest<VolunteerPage>(`/api/volunteers?${params}`);
}

export function getVolunteer(id: string) {
	return apiRequest<Volunteer>(`/api/volunteers/${encodeURIComponent(id)}`);
}

export function createVolunteer(input: VolunteerInput) {
	return apiRequest<Volunteer>('/api/volunteers', {
		method: 'POST',
		body: JSON.stringify(input)
	});
}

export function updateVolunteer(id: string, input: VolunteerInput) {
	return apiRequest<Volunteer>(`/api/volunteers/${encodeURIComponent(id)}`, {
		method: 'PUT',
		body: JSON.stringify(input)
	});
}

export function deleteVolunteer(id: string) {
	return apiRequest<void>(`/api/volunteers/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

export function bulkUpdateVolunteers(ids: string[], field: VolunteerBulkField, value: string) {
	return apiRequest<{ updated: number }>('/api/volunteers/bulk', {
		method: 'PATCH',
		body: JSON.stringify({ ids, field, value })
	});
}

export function bulkDeleteVolunteers(ids: string[]) {
	return apiRequest<{ deleted: number }>('/api/volunteers/bulk', {
		method: 'DELETE',
		body: JSON.stringify({ ids })
	});
}

export async function listDepartmentSuggestions(
	query: string,
	signal?: AbortSignal
): Promise<string[]> {
	const params = new URLSearchParams({ q: query, limit: '10' });
	const result = await apiRequest<{ departments: string[] }>(
		`/api/volunteer-options/departments?${params}`,
		{ signal }
	);
	return result.departments;
}
