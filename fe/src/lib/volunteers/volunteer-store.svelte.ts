import {
	createVolunteer,
	deleteVolunteer,
	getVolunteer,
	listVolunteers,
	updateVolunteer,
	type Volunteer,
	type VolunteerInput
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
	isLoading = $state(false);
	isSaving = $state(false);
	loaded = $state(false);
	lastLoadedAt = $state(0);
	private loadedQuery = '';
	private loadedStatus = '';
	private requestGeneration = 0;

	async load() {
		const generation = ++this.requestGeneration;
		const query = this.query;
		const status = this.status;
		const hasMatchingCache =
			this.loaded && this.loadedQuery === query && this.loadedStatus === status;
		this.isLoading = !hasMatchingCache;
		try {
			const items = await listVolunteers(query, status);
			if (generation !== this.requestGeneration) return;

			this.items = items;
			this.loaded = true;
			this.loadedQuery = query;
			this.loadedStatus = status;
			this.lastLoadedAt = Date.now();
		} catch (error) {
			if (generation !== this.requestGeneration) return;
			toastStore.error(message(error));
		} finally {
			if (generation === this.requestGeneration) this.isLoading = false;
		}
	}

	async refreshIfStale(ttlMs: number) {
		const cacheMatches =
			this.loaded && this.loadedQuery === this.query && this.loadedStatus === this.status;
		if (cacheMatches && Date.now() - this.lastLoadedAt < ttlMs) return;
		await this.load();
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
			toastStore.success(id ? 'Đã cập nhật hồ sơ' : 'Đã thêm hồ sơ công quả');
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
			this.items = this.items.filter((item) => item.id !== id);
			this.lastLoadedAt = Date.now();
			toastStore.success('Đã xoá hồ sơ');
			return true;
		} catch (error) {
			toastStore.error(message(error));
			return false;
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
