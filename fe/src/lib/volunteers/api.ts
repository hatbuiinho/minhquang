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

export async function listVolunteers(query = '', status = ''): Promise<Volunteer[]> {
	const params = new URLSearchParams();
	if (query) params.set('q', query);
	if (status) params.set('status', status);
	const result = await apiRequest<{ volunteers: Volunteer[] }>(`/api/volunteers?${params}`);
	return result.volunteers;
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
