import { Capacitor } from '@capacitor/core';

export const API_BASE_URL = apiBaseURL();

export type EventStatus = 'active' | 'archived' | 'cancelled';
export type ReminderJobStatus = 'pending' | 'processing' | 'sent' | 'cancelled' | 'failed';
export type DevicePlatform = 'android' | 'ios' | 'web';
export type AudienceType = 'self' | 'selected_users' | 'selected_groups' | 'all_users';
export type RecipientSourceType = 'self' | 'user' | 'group' | 'all_users';
export type ReminderImportance = 'normal' | 'urgent';

export type ReminderEvent = {
	id: string;
	user_id: string;
	title: string;
	description: string;
	starts_at: string;
	timezone: string;
	status: EventStatus;
	audience_type: AudienceType;
	reminder_generation: number;
	reminders: ReminderRule[];
	recipients?: EventRecipient[];
	created_at: string;
	updated_at: string;
};

export type ReminderRule = {
	id: string;
	event_id: string;
	offset_minutes: number;
	enabled: boolean;
	channel: 'push';
	importance: ReminderImportance;
	created_at: string;
	updated_at: string;
};

export type ReminderJob = {
	id: string;
	user_id: string;
	event_id: string;
	reminder_rule_id: string;
	event_title: string;
	event_starts_at: string;
	offset_minutes: number;
	channel: 'push';
	importance: ReminderImportance;
	status: ReminderJobStatus;
	scheduled_at: string;
	sent_at?: string | null;
	read_at?: string | null;
	dismissed_at?: string | null;
	snoozed_from_id?: string;
	snoozed_at?: string | null;
	cancelled_at?: string | null;
	reminder_generation: number;
	created_at: string;
	updated_at: string;
};

export type UserDevice = {
	id: string;
	user_id: string;
	platform: DevicePlatform;
	push_token: string;
	enabled: boolean;
	last_seen_at: string;
	created_at: string;
	updated_at: string;
};

export type EventRecipient = {
	event_id: string;
	user_id: string;
	source_type: RecipientSourceType;
	source_id?: string;
	created_at: string;
};

export type User = {
	id: string;
	name: string;
	email: string;
	active: boolean;
	created_at: string;
	updated_at: string;
};

export type Group = {
	id: string;
	name: string;
	description: string;
	active: boolean;
	created_at: string;
	updated_at: string;
};

export type AppUpdate = {
	available: boolean;
	platform?: string;
	channel?: string;
	version?: string;
	url?: string;
	checksum?: string;
	mandatory?: boolean;
	min_native_version?: string;
	max_native_version?: string;
	notes?: string;
};

export type ReminderRuleInput = {
	offset_minutes: number;
	enabled?: boolean;
	channel?: 'push';
	importance?: ReminderImportance;
};

export type EventInput = {
	title: string;
	description: string;
	starts_at: string;
	timezone: string;
	status?: EventStatus;
	reminders?: ReminderRuleInput[];
	audience_type?: AudienceType;
	recipient_user_ids?: string[];
	recipient_group_ids?: string[];
};

type EventListResponse = {
	events: ReminderEventResponse[];
	next_cursor?: string;
	has_more: boolean;
};

type UpcomingRemindersResponse = {
	reminders: ReminderJob[];
};

type ReminderJobsResponse = {
	reminders: ReminderJob[];
};

type UsersResponse = {
	users: User[];
};

type GroupsResponse = {
	groups: Group[];
};

type ReminderEventResponse = Omit<ReminderEvent, 'reminders'> & {
	reminders: ReminderRule[] | null;
};

type ApiErrorResponse = {
	error?: {
		code?: string;
		message?: string;
	};
};

const headers = {
	'Content-Type': 'application/json',
	'X-User-ID': 'local-user'
};

export type EventPage = {
	events: ReminderEvent[];
	next_cursor: string;
	has_more: boolean;
};

export async function listEvents(
	options: { cursor?: string; limit?: number } = {}
): Promise<EventPage> {
	const params = new URLSearchParams();
	params.set('limit', String(options.limit ?? 20));
	if (options.cursor) {
		params.set('cursor', options.cursor);
	}

	const response = await request<EventListResponse>(`/api/events?${params.toString()}`);
	return {
		events: response.events.map(normalizeEvent),
		next_cursor: response.next_cursor ?? '',
		has_more: response.has_more
	};
}

export async function getEvent(id: string): Promise<ReminderEvent> {
	return normalizeEvent(await request<ReminderEventResponse>(`/api/events/${id}`));
}

export async function createEvent(input: EventInput): Promise<ReminderEvent> {
	return normalizeEvent(
		await request<ReminderEventResponse>('/api/events', {
			method: 'POST',
			body: JSON.stringify(input)
		})
	);
}

export async function updateEvent(id: string, input: Partial<EventInput>): Promise<ReminderEvent> {
	return normalizeEvent(
		await request<ReminderEventResponse>(`/api/events/${id}`, {
			method: 'PATCH',
			body: JSON.stringify(input)
		})
	);
}

export async function deleteEvent(id: string): Promise<void> {
	await request<void>(`/api/events/${id}`, { method: 'DELETE' });
}

export async function listUpcomingReminders(limit = 50): Promise<ReminderJob[]> {
	const response = await request<UpcomingRemindersResponse>(
		`/api/reminders/upcoming?limit=${limit}`
	);
	return response.reminders;
}

export async function listReminderInbox(limit = 50): Promise<ReminderJob[]> {
	const response = await request<ReminderJobsResponse>(`/api/reminder-jobs?limit=${limit}`);
	return response.reminders;
}

export function markReminderJobRead(id: string): Promise<ReminderJob> {
	return request<ReminderJob>(`/api/reminder-jobs/${encodeURIComponent(id)}/read`, {
		method: 'POST'
	});
}

export function dismissReminderJob(id: string): Promise<ReminderJob> {
	return request<ReminderJob>(`/api/reminder-jobs/${encodeURIComponent(id)}/dismiss`, {
		method: 'POST'
	});
}

export function snoozeReminderJob(id: string, delayMinutes: number): Promise<ReminderJob> {
	return request<ReminderJob>(`/api/reminder-jobs/${encodeURIComponent(id)}/snooze`, {
		method: 'POST',
		body: JSON.stringify({ delay_minutes: delayMinutes })
	});
}

export function registerDevice(input: {
	platform: DevicePlatform;
	push_token: string;
}): Promise<UserDevice> {
	return request<UserDevice>('/api/devices', {
		method: 'POST',
		body: JSON.stringify(input)
	});
}

export async function listUsers(): Promise<User[]> {
	const response = await request<UsersResponse>('/api/users');
	return response.users;
}

export async function listGroups(): Promise<Group[]> {
	const response = await request<GroupsResponse>('/api/groups');
	return response.groups;
}

export function getLatestAndroidUpdate(input: {
	channel: string;
	current_version: string;
	native_version: string;
}): Promise<AppUpdate> {
	const params = new URLSearchParams({
		channel: input.channel,
		current_version: input.current_version,
		native_version: input.native_version
	});

	return request<AppUpdate>(`/api/app-updates/android/latest?${params.toString()}`);
}

function normalizeEvent(event: ReminderEventResponse): ReminderEvent {
	return {
		...event,
		reminders: event.reminders ?? []
	};
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
	const response = await fetch(`${API_BASE_URL}${path}`, {
		...init,
		headers: {
			...headers,
			...init.headers
		}
	});

	if (!response.ok) {
		throw new Error(await errorMessage(response));
	}

	if (response.status === 204) {
		return undefined as T;
	}

	return response.json() as Promise<T>;
}

async function errorMessage(response: Response): Promise<string> {
	try {
		const body = (await response.json()) as ApiErrorResponse;
		return body.error?.message ?? `Request failed with ${response.status}`;
	} catch {
		return `Request failed with ${response.status}`;
	}
}

function apiBaseURL(): string {
	const configuredURL = import.meta.env.VITE_API_BASE_URL;
	if (configuredURL) {
		return configuredURL;
	}

	if (Capacitor.getPlatform() === 'android') {
		return 'http://10.0.2.2:8080';
	}

	if (typeof window === 'undefined') {
		return 'http://localhost:8080';
	}

	return `${window.location.protocol}//${window.location.hostname}:8080`;
}
