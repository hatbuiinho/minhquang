<script lang="ts">
	import { onMount } from 'svelte';
	import { departmentStore, type DepartmentFilter } from '$lib/departments/department-store.svelte';
	import type { Department } from '$lib/departments/api';
	import { popupStore } from '$lib/ui/popup-store.svelte';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';

	let showCreate = $state(false);
	let createName = $state('');
	let editingID = $state('');
	let editingName = $state('');
	let searchTimer: ReturnType<typeof setTimeout> | undefined;
	const filters: { value: DepartmentFilter; label: string }[] = [
		{ value: 'all', label: 'Tất cả' },
		{ value: 'true', label: 'Đang dùng' },
		{ value: 'false', label: 'Đã ẩn' }
	];

	onMount(() => {
		void departmentStore.refreshIfStale(45_000);
		return () => {
			if (searchTimer) clearTimeout(searchTimer);
		};
	});

	function search() {
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => void departmentStore.load(), 350);
	}

	function changeFilter(filter: DepartmentFilter) {
		if (departmentStore.filter === filter) return;
		departmentStore.filter = filter;
		void departmentStore.refreshIfStale(45_000);
	}

	async function create(event: SubmitEvent) {
		event.preventDefault();
		if (!(await departmentStore.create(createName))) return;
		createName = '';
		showCreate = false;
	}

	function beginEdit(item: Department) {
		editingID = item.id;
		editingName = item.name;
	}

	async function rename(event: SubmitEvent) {
		event.preventDefault();
		if (!(await departmentStore.rename(editingID, editingName))) return;
		editingID = '';
		editingName = '';
	}

	async function toggle(item: Department) {
		const action = item.active ? 'ngừng sử dụng' : 'mở lại';
		const confirmed = await popupStore.confirm({
			title: item.active ? 'Ẩn phân ban?' : 'Mở lại phân ban?',
			message: `Bạn có chắc muốn ${action} ${item.name}?`,
			confirmLabel: item.active ? 'Ngừng sử dụng' : 'Mở lại'
		});
		if (confirmed) await departmentStore.setActive(item.id, !item.active);
	}

	async function remove(item: Department) {
		const confirmed = await popupStore.confirm({
			title: 'Xoá phân ban?',
			message:
				item.volunteer_count > 0
					? `${item.name} đang có ${item.volunteer_count} hồ sơ nên không thể xoá.`
					: `Phân ban ${item.name} sẽ bị xoá vĩnh viễn.`,
			confirmLabel: 'Xoá',
			tone: 'danger'
		});
		if (confirmed && item.volunteer_count === 0) await departmentStore.remove(item.id);
	}
</script>

<section class="h-full overflow-y-auto px-4 py-4">
	<div class="mb-4 flex items-center justify-between gap-3">
		<div>
			<p class="text-sm font-medium">Danh mục phân ban</p>
			<p class="mt-0.5 text-xs text-[var(--color-text-secondary)]">
				{departmentStore.items.length} phân ban
			</p>
		</div>
		<button
			type="button"
			class="grid h-10 w-10 place-items-center rounded-md bg-[var(--color-primary)] text-white"
			aria-label={showCreate ? 'Đóng biểu mẫu' : 'Thêm phân ban'}
			onclick={() => (showCreate = !showCreate)}
		>
			<span class={showCreate ? 'icon-[lucide--x] h-5 w-5' : 'icon-[lucide--plus] h-5 w-5'} aria-hidden="true"></span>
		</button>
	</div>

	{#if showCreate}
		<form class="mb-4 flex gap-2" onsubmit={create}>
			<input
				bind:value={createName}
				required
				maxlength="60"
				placeholder="Tên phân ban"
				aria-label="Tên phân ban mới"
				class="h-11 min-w-0 flex-1 rounded-md border-[var(--color-border-strong)]"
			/>
			<button
				type="submit"
				disabled={departmentStore.isSaving || !createName.trim()}
				class="grid h-11 w-11 place-items-center rounded-md bg-[var(--color-primary)] text-white disabled:opacity-50"
				aria-label="Lưu phân ban"
			><span class="icon-[lucide--check] h-5 w-5" aria-hidden="true"></span></button>
		</form>
	{/if}

	<div class="relative mb-3">
		<span class="icon-[lucide--search] pointer-events-none absolute top-3.5 left-3 h-4 w-4 text-[var(--color-text-muted)]" aria-hidden="true"></span>
		<input
			bind:value={departmentStore.query}
			type="search"
			placeholder="Tìm phân ban"
			aria-label="Tìm phân ban"
			class="h-11 w-full rounded-md border-[var(--color-border-strong)] pl-9"
			oninput={search}
		/>
	</div>

	<div class="mb-4 grid grid-cols-3 rounded-md bg-[var(--color-surface-muted)] p-1">
		{#each filters as filter (filter.value)}
			<button
				type="button"
				class={[
					'h-9 rounded text-xs font-semibold',
					departmentStore.filter === filter.value
						? 'bg-[var(--color-surface)] text-[var(--color-primary-dark)] shadow-sm'
						: 'text-[var(--color-text-secondary)]'
				]}
				onclick={() => changeFilter(filter.value)}>{filter.label}</button>
		{/each}
	</div>

	{#if departmentStore.isLoading && departmentStore.items.length === 0}
		<div class="py-16"><LoadingIndicator label="Đang tải phân ban..." /></div>
	{:else if departmentStore.items.length === 0}
		<div class="py-16 text-center">
			<span class="icon-[lucide--inbox] mx-auto block h-8 w-8 text-[var(--color-text-muted)]" aria-hidden="true"></span>
			<p class="mt-2 text-sm text-[var(--color-text-secondary)]">Không có phân ban phù hợp</p>
		</div>
	{:else}
		<ul class="divide-y divide-[var(--color-border)] border-y border-[var(--color-border)]">
			{#each departmentStore.items as item (item.id)}
				<li class="py-3">
					{#if editingID === item.id}
						<form class="flex gap-2" onsubmit={rename}>
							<input bind:value={editingName} required maxlength="60" aria-label="Tên phân ban" class="h-10 min-w-0 flex-1 rounded-md border-[var(--color-border-strong)]" />
							<button type="submit" disabled={departmentStore.isSaving || !editingName.trim()} class="grid h-10 w-10 place-items-center rounded-md bg-[var(--color-primary)] text-white disabled:opacity-50" aria-label="Lưu tên"><span class="icon-[lucide--check] h-4 w-4" aria-hidden="true"></span></button>
							<button type="button" class="grid h-10 w-10 place-items-center rounded-md border border-[var(--color-border-strong)]" aria-label="Huỷ sửa" onclick={() => (editingID = '')}><span class="icon-[lucide--x] h-4 w-4" aria-hidden="true"></span></button>
						</form>
					{:else}
						<div class="flex items-center gap-3">
							<div class="min-w-0 flex-1">
								<div class="flex items-center gap-2">
									<p class="truncate text-sm font-semibold">{item.name}</p>
									<span class={['shrink-0 rounded px-1.5 py-0.5 text-[10px] font-semibold', item.active ? 'bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]' : 'bg-[var(--color-surface-muted)] text-[var(--color-text-muted)]']}>{item.active ? 'Đang dùng' : 'Đã ẩn'}</span>
								</div>
								<p class="mt-1 text-xs text-[var(--color-text-secondary)]">{item.volunteer_count} hồ sơ</p>
							</div>
							<button type="button" class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-text-secondary)]" aria-label="Sửa tên" title="Sửa tên" onclick={() => beginEdit(item)}><span class="icon-[lucide--pencil] h-4 w-4" aria-hidden="true"></span></button>
							<button type="button" class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-text-secondary)]" aria-label={item.active ? 'Ngừng sử dụng' : 'Mở lại'} title={item.active ? 'Ngừng sử dụng' : 'Mở lại'} onclick={() => void toggle(item)}><span class={item.active ? 'icon-[lucide--eye-off] h-4 w-4' : 'icon-[lucide--eye] h-4 w-4'} aria-hidden="true"></span></button>
							<button type="button" disabled={item.volunteer_count > 0} class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-danger)] disabled:opacity-30" aria-label="Xoá phân ban" title={item.volunteer_count > 0 ? 'Không thể xoá phân ban đang có hồ sơ' : 'Xoá phân ban'} onclick={() => void remove(item)}><span class="icon-[lucide--trash-2] h-4 w-4" aria-hidden="true"></span></button>
						</div>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</section>
