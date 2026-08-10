import { SvelteMap } from 'svelte/reactivity';

import {
	createEvent,
	deleteEvent,
	getEvent,
	listGroups,
	listEvents,
	listUsers,
	updateEvent,
	type AudienceType,
	type EventInput,
	type EventStatus,
	type Group,
	type ReminderImportance,
	type ReminderEvent,
	type User
} from './api';
import { fromDateTimeLocal, toDateTimeLocal } from './time';
import { toastStore } from '$lib/ui/toast-store.svelte';

export type EventFormState = {
	title: string;
	description: string;
	starts_at: string;
	timezone: string;
	status: EventStatus;
	reminder_offsets: number[];
	reminder_importance: ReminderImportance;
	audience_type: AudienceType;
	recipient_user_ids: string[];
	recipient_group_ids: string[];
};

export const reminderPresetOptions = [
	{ label: 'Đúng thời điểm sự kiện', offsetMinutes: 0 },
	{ label: 'Trước 10 phút', offsetMinutes: 10 },
	{ label: 'Trước 1 giờ', offsetMinutes: 60 },
	{ label: 'Trước 1 ngày', offsetMinutes: 1440 },
	{ label: 'Trước 3 ngày', offsetMinutes: 4320 },
	{ label: 'Trước 7 ngày', offsetMinutes: 10080 }
];

class EventStore {
	eventsById = new SvelteMap<string, ReminderEvent>();
	eventIndex = new SvelteMap<string, number>();
	loadingEventIds = new SvelteMap<string, boolean>();
	deletedEventIds = new Set<string>();
	eventIds = $state<string[]>([]);
	form = $state<EventFormState>(this.emptyForm());
	loaded = $state(false);
	isLoading = $state(false);
	isLoadingMore = $state(false);
	hasMore = $state(false);
	nextCursor = $state('');
	isSaving = $state(false);
	isLoadingAudience = $state(false);
	error = $state('');
	editingId = $state<string | null>(null);
	users = $state<User[]>([]);
	groups = $state<Group[]>([]);
	audienceLoaded = $state(false);
	lastLoadedAt = $state(0);

	async loadOnce() {
		if (this.loaded || this.isLoading) return;

		await this.refresh();
	}

	async refreshIfStale(ttlMs: number) {
		if (!this.loaded || Date.now() - this.lastLoadedAt >= ttlMs) {
			await this.refresh();
		}
	}

	async refresh() {
		if (this.isLoading) return;

		this.isLoading = true;
		this.error = '';

		try {
			const page = await listEvents();
			this.replaceFirstPage(page.events);
			this.nextCursor = page.next_cursor;
			this.hasMore = page.has_more;
			this.loaded = true;
			this.lastLoadedAt = Date.now();
		} catch (err) {
			this.error = messageFromError(err);
			toastStore.error(this.error);
		} finally {
			this.isLoading = false;
		}
	}

	async loadMore() {
		if (!this.loaded || !this.hasMore || this.isLoadingMore) return;

		this.isLoadingMore = true;
		this.error = '';

		try {
			const page = await listEvents({ cursor: this.nextCursor });
			this.appendPage(page.events);
			this.nextCursor = page.next_cursor;
			this.hasMore = page.has_more;
		} catch (err) {
			this.error = messageFromError(err);
			toastStore.error(this.error);
		} finally {
			this.isLoadingMore = false;
		}
	}

	get(id: string | undefined): ReminderEvent | undefined {
		return id ? this.eventsById.get(id) : undefined;
	}

	isEventLoading(id: string | undefined): boolean {
		return id ? this.loadingEventIds.has(id) : false;
	}

	async loadById(id: string | undefined) {
		if (
			!id ||
			this.eventsById.has(id) ||
			this.loadingEventIds.has(id) ||
			this.deletedEventIds.has(id)
		) {
			return;
		}

		this.loadingEventIds.set(id, true);
		this.error = '';

		try {
			const event = await getEvent(id);
			this.deletedEventIds.delete(id);
			this.upsert(event);
		} catch (err) {
			this.error = messageFromError(err);
			toastStore.error(this.error);
		} finally {
			this.loadingEventIds.delete(id);
		}
	}

	orderedEvents(): ReminderEvent[] {
		const items: ReminderEvent[] = [];
		for (const id of this.eventIds) {
			const event = this.eventsById.get(id);
			if (event) items.push(event);
		}
		return items;
	}

	prepareCreate() {
		this.editingId = null;
		this.form = this.emptyForm();
		this.error = '';
		void this.loadAudienceOptions();
	}

	prepareEdit(eventId: string) {
		const event = this.eventsById.get(eventId);
		if (!event) {
			this.error = 'Không tìm thấy sự kiện';
			toastStore.error(this.error);
			return;
		}

		this.editingId = event.id;
		this.form = {
			title: event.title,
			description: event.description,
			starts_at: toDateTimeLocal(event.starts_at),
			timezone: event.timezone,
			status: event.status,
			reminder_offsets: event.reminders.map((rule) => rule.offset_minutes),
			reminder_importance: event.reminders.some((rule) => rule.importance === 'urgent')
				? 'urgent'
				: 'normal',
			audience_type: event.audience_type ?? 'self',
			recipient_user_ids: (event.recipients ?? [])
				.filter((recipient) => recipient.source_type === 'user')
				.map((recipient) => recipient.user_id),
			recipient_group_ids: [
				...new Set(
					(event.recipients ?? [])
						.filter((recipient) => recipient.source_type === 'group' && recipient.source_id)
						.map((recipient) => recipient.source_id as string)
				)
			]
		};
		this.error = '';
		void this.loadAudienceOptions();
	}

	toggleReminderOffset(offset: number) {
		const exists = this.form.reminder_offsets.includes(offset);
		this.form.reminder_offsets = exists
			? this.form.reminder_offsets.filter((value) => value !== offset)
			: [...this.form.reminder_offsets, offset].sort((left, right) => left - right);
	}

	toggleRecipientUser(userId: string) {
		const exists = this.form.recipient_user_ids.includes(userId);
		this.form.recipient_user_ids = exists
			? this.form.recipient_user_ids.filter((id) => id !== userId)
			: [...this.form.recipient_user_ids, userId];
	}

	toggleRecipientGroup(groupId: string) {
		const exists = this.form.recipient_group_ids.includes(groupId);
		this.form.recipient_group_ids = exists
			? this.form.recipient_group_ids.filter((id) => id !== groupId)
			: [...this.form.recipient_group_ids, groupId];
	}

	async loadAudienceOptions() {
		if (this.audienceLoaded || this.isLoadingAudience) return;

		this.isLoadingAudience = true;
		try {
			const [users, groups] = await Promise.all([listUsers(), listGroups()]);
			this.users = users;
			this.groups = groups;
			this.audienceLoaded = true;
		} catch (err) {
			this.error = messageFromError(err);
			toastStore.error(this.error);
		} finally {
			this.isLoadingAudience = false;
		}
	}

	async saveForm(): Promise<ReminderEvent | undefined> {
		if (this.isSaving) return undefined;

		this.isSaving = true;
		this.error = '';

		const input: EventInput = {
			title: this.form.title,
			description: this.form.description,
			starts_at: fromDateTimeLocal(this.form.starts_at),
			timezone: this.form.timezone,
			status: this.form.status,
			reminders: [...this.form.reminder_offsets]
				.sort((left, right) => left - right)
				.map((offset) => ({
					offset_minutes: offset,
					importance: this.form.reminder_importance
				})),
			audience_type: this.form.audience_type,
			recipient_user_ids:
				this.form.audience_type === 'selected_users' ? this.form.recipient_user_ids : [],
			recipient_group_ids:
				this.form.audience_type === 'selected_groups' ? this.form.recipient_group_ids : []
		};

		try {
			const event = this.editingId
				? await updateEvent(this.editingId, input)
				: await createEvent(input);
			toastStore.success(this.editingId ? 'Đã lưu sự kiện' : 'Đã tạo sự kiện');
			this.upsert(event, !this.editingId);
			this.lastLoadedAt = Date.now();
			this.prepareCreate();
			return event;
		} catch (err) {
			this.error = messageFromError(err);
			toastStore.error(this.error);
			return undefined;
		} finally {
			this.isSaving = false;
		}
	}

	async remove(id: string): Promise<boolean> {
		this.error = '';

		try {
			await deleteEvent(id);
			this.deletedEventIds.add(id);
			this.removeLocal(id);
			this.lastLoadedAt = Date.now();
			if (this.editingId === id) {
				this.prepareCreate();
			}
			toastStore.success('Đã xoá sự kiện');
			return true;
		} catch (err) {
			this.error = messageFromError(err);
			toastStore.error(this.error);
			return false;
		}
	}

	private upsert(event: ReminderEvent, prepend = false) {
		this.deletedEventIds.delete(event.id);
		this.eventsById.set(event.id, event);
		if (this.eventIndex.has(event.id)) return;

		if (!prepend) {
			this.eventIndex.set(event.id, this.eventIds.length);
			this.eventIds = [...this.eventIds, event.id];
			return;
		}

		this.reindex([event.id, ...this.eventIds]);
	}

	private appendPage(events: ReminderEvent[]) {
		const nextIds = [...this.eventIds];
		for (const event of events) {
			this.eventsById.set(event.id, event);
			if (this.eventIndex.has(event.id)) continue;

			this.eventIndex.set(event.id, nextIds.length);
			nextIds.push(event.id);
		}
		this.eventIds = nextIds;
	}

	private replaceFirstPage(events: ReminderEvent[]) {
		for (const event of events) {
			this.deletedEventIds.delete(event.id);
			this.eventsById.set(event.id, event);
		}
		this.reindex(events.map((event) => event.id));
	}

	private removeLocal(id: string) {
		const index = this.eventIndex.get(id);
		this.eventsById.delete(id);
		if (index === undefined) return;

		this.reindex(this.eventIds.filter((eventId) => eventId !== id));
	}

	private reindex(ids: string[]) {
		this.eventIndex.clear();
		ids.forEach((id, index) => this.eventIndex.set(id, index));
		this.eventIds = ids;
	}

	private emptyForm(): EventFormState {
		return {
			title: '',
			description: '',
			starts_at: toDateTimeLocal(new Date(Date.now() + 60 * 60 * 1000).toISOString()),
			timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC',
			status: 'active',
			reminder_offsets: [],
			reminder_importance: 'normal',
			audience_type: 'self',
			recipient_user_ids: [],
			recipient_group_ids: []
		};
	}
}

function messageFromError(err: unknown): string {
	return err instanceof Error ? err.message : 'Đã có lỗi xảy ra';
}

export const eventStore = new EventStore();
