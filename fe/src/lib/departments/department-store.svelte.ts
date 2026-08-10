import {
	createDepartment,
	deleteDepartment,
	listDepartments,
	setDepartmentActive,
	updateDepartment,
	type Department
} from './api';
import { toastStore } from '$lib/ui/toast-store.svelte';

export type DepartmentFilter = 'all' | 'true' | 'false';

class DepartmentStore {
	items = $state<Department[]>([]);
	query = $state('');
	filter = $state<DepartmentFilter>('all');
	isLoading = $state(false);
	isSaving = $state(false);
	loaded = $state(false);
	lastLoadedAt = $state(0);
	private loadedQuery = '';
	private loadedFilter: DepartmentFilter = 'all';
	private requestGeneration = 0;

	async load() {
		const generation = ++this.requestGeneration;
		const query = this.query;
		const filter = this.filter;
		const cacheMatches =
			this.loaded && this.loadedQuery === query && this.loadedFilter === filter;
		this.isLoading = !cacheMatches;
		try {
			const items = await listDepartments(query, filter);
			if (generation !== this.requestGeneration) return;
			this.items = items;
			this.loaded = true;
			this.loadedQuery = query;
			this.loadedFilter = filter;
			this.lastLoadedAt = Date.now();
		} catch (error) {
			if (generation === this.requestGeneration) toastStore.error(message(error));
		} finally {
			if (generation === this.requestGeneration) this.isLoading = false;
		}
	}

	async refreshIfStale(ttlMs: number) {
		const matches =
			this.loaded && this.loadedQuery === this.query && this.loadedFilter === this.filter;
		if (matches && Date.now() - this.lastLoadedAt < ttlMs) return;
		await this.load();
	}

	async create(name: string) {
		return this.mutate(() => createDepartment(name), 'Đã thêm phân ban');
	}

	async rename(id: string, name: string) {
		return this.mutate(() => updateDepartment(id, name), 'Đã đổi tên phân ban');
	}

	async setActive(id: string, active: boolean) {
		return this.mutate(
			() => setDepartmentActive(id, active),
			active ? 'Đã mở lại phân ban' : 'Đã ngừng sử dụng phân ban'
		);
	}

	async remove(id: string) {
		if (this.isSaving) return false;
		this.isSaving = true;
		try {
			await deleteDepartment(id);
			this.items = this.items.filter((item) => item.id !== id);
			this.lastLoadedAt = Date.now();
			toastStore.success('Đã xoá phân ban');
			return true;
		} catch (error) {
			toastStore.error(message(error));
			return false;
		} finally {
			this.isSaving = false;
		}
	}

	private async mutate(request: () => Promise<Department>, successMessage: string) {
		if (this.isSaving) return false;
		this.isSaving = true;
		try {
			await request();
			toastStore.success(successMessage);
			await this.load();
			return true;
		} catch (error) {
			toastStore.error(message(error));
			return false;
		} finally {
			this.isSaving = false;
		}
	}
}

function message(error: unknown) {
	return error instanceof Error ? error.message : 'Có lỗi xảy ra';
}

export const departmentStore = new DepartmentStore();
