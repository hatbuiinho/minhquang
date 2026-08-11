import {
	bulkDeleteVolunteers,
	bulkUpdateVolunteers,
	createVolunteer,
	deleteVolunteer,
	getVolunteer,
	listVolunteers,
	updateVolunteer,
	type Volunteer,
	type VolunteerInput,
	type VolunteerBulkField,
	type VolunteerSortKey
} from './api';
import { toastStore } from '$lib/ui/toast-store.svelte';

export type VolunteerForm = {
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

class VolunteerStore {
	items = $state<Volunteer[]>([]);
	selected = $state<Volunteer | null>(null);
	form = $state<VolunteerForm>(this.emptyForm());
	query = $state('');
	status = $state('active');
	departmentId = $state('');
	departmentName = $state('');
	sortKey = $state<VolunteerSortKey>('arrival_date');
	sortDirection = $state<'asc' | 'desc'>('desc');
	total = $state(0);
	hasMore = $state(false);
	isLoading = $state(false);
	isRefreshing = $state(false);
	isSorting = $state(false);
	isLoadingMore = $state(false);
	isSaving = $state(false);
	isBulkSaving = $state(false);
	isAvatarSaving = $state(false);
	loaded = $state(false);
	lastLoadedAt = $state(0);
	private loadedQuery = '';
	private loadedStatus = '';
	private loadedDepartmentId = '';
	private loadedSortKey: VolunteerSortKey = 'arrival_date';
	private loadedSortDirection: 'asc' | 'desc' = 'desc';
	private requestGeneration = 0;
	private readonly pageSize = 20;

	async load(limit = this.pageSize, preserveContent = this.loaded, sorting = false) {
		const generation = ++this.requestGeneration;
		const query = this.query;
		const status = this.status;
		const departmentId = this.departmentId;
		const hasMatchingCache =
			this.loaded &&
			this.loadedQuery === query &&
			this.loadedStatus === status &&
			this.loadedDepartmentId === departmentId &&
			this.loadedSortKey === this.sortKey &&
			this.loadedSortDirection === this.sortDirection;
		const keepContent = preserveContent && this.loaded;
		this.isLoading = !keepContent && !hasMatchingCache;
		this.isRefreshing = keepContent;
		this.isSorting = keepContent && sorting;
		try {
			const page = await listVolunteers(
				query,
				status,
				departmentId,
				limit,
				0,
				this.sortKey,
				this.sortDirection
			);
			if (generation !== this.requestGeneration) return;

			this.items = page.volunteers;
			this.total = page.total;
			this.hasMore = page.has_more;
			this.loaded = true;
			this.loadedQuery = query;
			this.loadedStatus = status;
			this.loadedDepartmentId = departmentId;
			this.loadedSortKey = this.sortKey;
			this.loadedSortDirection = this.sortDirection;
			this.lastLoadedAt = Date.now();
		} catch (error) {
			if (generation !== this.requestGeneration) return;
			if (sorting) {
				this.sortKey = this.loadedSortKey;
				this.sortDirection = this.loadedSortDirection;
			}
			toastStore.error(message(error));
		} finally {
			if (generation === this.requestGeneration) {
				this.isLoading = false;
				this.isRefreshing = false;
				this.isSorting = false;
			}
		}
	}

	async loadSorted() {
		await this.load(this.pageSize, true, true);
	}

	async loadMore() {
		if (this.isLoading || this.isRefreshing || this.isLoadingMore || !this.hasMore) return;
		const generation = this.requestGeneration;
		const query = this.query;
		const status = this.status;
		const departmentId = this.departmentId;
		const sortKey = this.sortKey;
		const sortDirection = this.sortDirection;
		this.isLoadingMore = true;
		try {
			const page = await listVolunteers(
				query,
				status,
				departmentId,
				this.pageSize,
				this.items.length,
				sortKey,
				sortDirection
			);
			if (
				generation !== this.requestGeneration ||
				query !== this.query ||
				status !== this.status ||
				departmentId !== this.departmentId ||
				sortKey !== this.sortKey ||
				sortDirection !== this.sortDirection
			)
				return;
			const existingIDs = new Set(this.items.map((item) => item.id));
			this.items = [...this.items, ...page.volunteers.filter((item) => !existingIDs.has(item.id))];
			this.total = page.total;
			this.hasMore = page.has_more;
			this.lastLoadedAt = Date.now();
		} catch (error) {
			toastStore.error(message(error));
		} finally {
			this.isLoadingMore = false;
		}
	}

	async refreshIfStale(ttlMs: number) {
		const cacheMatches =
			this.loaded &&
			this.loadedQuery === this.query &&
			this.loadedStatus === this.status &&
			this.loadedDepartmentId === this.departmentId &&
			this.loadedSortKey === this.sortKey &&
			this.loadedSortDirection === this.sortDirection;
		if (cacheMatches && Date.now() - this.lastLoadedAt < ttlMs) return;
		await this.load(Math.max(this.pageSize, this.items.length));
	}

	filterByDepartment(id: string, name: string) {
		this.departmentId = id;
		this.departmentName = name;
		this.query = '';
		this.status = 'active';
	}

	prepareCreate() {
		this.selected = null;
		this.form = this.emptyForm();
	}

	async prepareEdit(id: string) {
		const item = this.items.find((candidate) => candidate.id === id) ?? (await this.fetch(id));
		if (!item) return;
		this.selected = item;
		this.form = {
			full_name: item.full_name,
			dharma_name: item.dharma_name,
			birth_date: item.birth_date,
			cultivation_place: item.cultivation_place,
			phone: item.phone,
			department: item.department,
			notes: item.notes,
			avatar_url: item.avatar_url,
			arrival_date: dateValue(item.arrival_date),
			departure_date: item.departure_date ? dateValue(item.departure_date) : ''
		};
	}

	async fetch(id: string): Promise<Volunteer | null> {
		try {
			const item = await getVolunteer(id);
			this.selected = item;
			return item;
		} catch (error) {
			toastStore.error(message(error));
			return null;
		}
	}

	async save(id?: string): Promise<Volunteer | null> {
		if (this.isSaving) return null;
		this.isSaving = true;
		const input: VolunteerInput = { ...this.form };
		try {
			const item = id ? await updateVolunteer(id, input) : await createVolunteer(input);
			toastStore.success(id ? 'Đã cập nhật Huynh đệ' : 'Đã thêm Huynh đệ công quả');
			await this.load();
			return item;
		} catch (error) {
			toastStore.error(message(error));
			return null;
		} finally {
			this.isSaving = false;
		}
	}

	async remove(id: string): Promise<boolean> {
		try {
			await deleteVolunteer(id);
			await this.load(Math.max(this.pageSize, this.items.length));
			toastStore.success('Đã xoá Huynh đệ');
			return true;
		} catch (error) {
			toastStore.error(message(error));
			return false;
		}
	}

	async updateAvatar(id: string, avatarURL: string): Promise<Volunteer | null> {
		if (this.isAvatarSaving) return null;
		this.isAvatarSaving = true;
		try {
			await bulkUpdateVolunteers([id], 'avatar_url', avatarURL);
			const item = await getVolunteer(id);
			this.selected = item;
			this.items = this.items.map((candidate) => (candidate.id === id ? item : candidate));
			toastStore.success('Đã cập nhật ảnh đại diện');
			return item;
		} catch (error) {
			toastStore.error(message(error));
			return null;
		} finally {
			this.isAvatarSaving = false;
		}
	}

	async bulkUpdate(
		ids: string[],
		field: VolunteerBulkField,
		value: string
	): Promise<number | null> {
		if (this.isBulkSaving) return null;
		this.isBulkSaving = true;
		try {
			const result = await bulkUpdateVolunteers(ids, field, value);
			await this.load(Math.max(this.pageSize, this.items.length));
			toastStore.success(`Đã cập nhật ${result.updated} Huynh đệ`);
			return result.updated;
		} catch (error) {
			toastStore.error(message(error));
			return null;
		} finally {
			this.isBulkSaving = false;
		}
	}

	async bulkDelete(ids: string[]): Promise<number | null> {
		if (this.isBulkSaving) return null;
		this.isBulkSaving = true;
		try {
			const result = await bulkDeleteVolunteers(ids);
			await this.load(Math.max(this.pageSize, this.items.length));
			toastStore.success(`Đã xoá ${result.deleted} Huynh đệ`);
			return result.deleted;
		} catch (error) {
			toastStore.error(message(error));
			return null;
		} finally {
			this.isBulkSaving = false;
		}
	}

	private emptyForm(): VolunteerForm {
		return {
			full_name: '',
			dharma_name: '',
			birth_date: '',
			cultivation_place: '',
			phone: '',
			department: '',
			notes: '',
			avatar_url: '',
			arrival_date: new Date().toISOString().slice(0, 10),
			departure_date: ''
		};
	}
}

function dateValue(value: string) {
	return value.slice(0, 10);
}
function message(error: unknown) {
	return error instanceof Error ? error.message : 'Có lỗi xảy ra';
}

export const volunteerStore = new VolunteerStore();
