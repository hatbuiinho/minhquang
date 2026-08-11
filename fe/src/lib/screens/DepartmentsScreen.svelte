<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { departmentStore, type DepartmentFilter } from '$lib/departments/department-store.svelte';
	import type { Department } from '$lib/departments/api';
	import { popupStore } from '$lib/ui/popup-store.svelte';
	import { router } from '$lib/navigation/router.svelte';
	import { volunteerStore } from '$lib/volunteers/volunteer-store.svelte';
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

	async function toggleCreate() {
		showCreate = !showCreate;
		if (!showCreate) return;
		editingID = '';
		await focusVisibleInput('[data-department-create-input]');
	}

	async function beginEdit(item: Department) {
		showCreate = false;
		editingID = item.id;
		editingName = item.name;
		await focusVisibleInput('[data-department-edit-input]');
	}

	async function focusVisibleInput(selector: string) {
		await tick();
		const input = [...document.querySelectorAll<HTMLInputElement>(selector)].find(
			(candidate) => candidate.offsetParent !== null
		);
		input?.focus();
		const end = input?.value.length ?? 0;
		input?.setSelectionRange(end, end);
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
					? `${item.name} đang có ${item.volunteer_count} Huynh đệ nên không thể xoá.`
					: `Phân ban ${item.name} sẽ bị xoá vĩnh viễn.`,
			confirmLabel: 'Xoá',
			tone: 'danger'
		});
		if (confirmed && item.volunteer_count === 0) await departmentStore.remove(item.id);
	}

	function openVolunteerList(item: Department) {
		volunteerStore.filterByDepartment(item.id, item.name);
		router.openMain('volunteers');
	}
</script>

<section class="h-full overflow-y-auto px-4 py-4 md:px-6 md:py-6 lg:px-8">
	<div class="mx-auto max-w-[1200px]">
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
				onclick={toggleCreate}
			>
				<span
					class={showCreate ? 'icon-[lucide--x] h-5 w-5' : 'icon-[lucide--plus] h-5 w-5'}
					aria-hidden="true"
				></span>
			</button>
		</div>

		{#if showCreate}
			<form class="mb-4 flex max-w-xl gap-2" onsubmit={create}>
				<input
					data-department-create-input
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
					><span class="icon-[lucide--check] h-5 w-5" aria-hidden="true"></span></button
				>
			</form>
		{/if}

		<div class="mb-4 flex flex-col gap-3 md:flex-row md:items-center">
			<div class="relative md:min-w-0 md:flex-1">
				<span
					class="pointer-events-none absolute top-3.5 left-3 icon-[lucide--search] h-4 w-4 text-[var(--color-text-muted)]"
					aria-hidden="true"
				></span>
				<input
					bind:value={departmentStore.query}
					type="search"
					placeholder="Tìm phân ban"
					aria-label="Tìm phân ban"
					class="h-11 w-full rounded-md border-[var(--color-border-strong)] pl-9"
					oninput={search}
				/>
			</div>

			<div
				class="grid grid-cols-3 rounded-md bg-[var(--color-surface-muted)] p-1 md:w-[320px] md:shrink-0"
			>
				{#each filters as filter (filter.value)}
					<button
						type="button"
						class={[
							'h-9 rounded text-xs font-semibold',
							departmentStore.filter === filter.value
								? 'bg-[var(--color-surface)] text-[var(--color-primary-dark)] shadow-sm'
								: 'text-[var(--color-text-secondary)]'
						]}
						onclick={() => changeFilter(filter.value)}>{filter.label}</button
					>
				{/each}
			</div>
		</div>

		{#if departmentStore.isLoading && departmentStore.items.length === 0}
			<div class="py-16"><LoadingIndicator label="Đang tải phân ban..." /></div>
		{:else if departmentStore.items.length === 0}
			<div class="py-16 text-center">
				<span
					class="mx-auto icon-[lucide--inbox] block h-8 w-8 text-[var(--color-text-muted)]"
					aria-hidden="true"
				></span>
				<p class="mt-2 text-sm text-[var(--color-text-secondary)]">Không có phân ban phù hợp</p>
			</div>
		{:else}
			<ul
				class="divide-y divide-[var(--color-border)] border-y border-[var(--color-border)] md:hidden"
			>
				{#each departmentStore.items as item (item.id)}
					<li class="py-3">
						{#if editingID === item.id}
							<form class="flex gap-2" onsubmit={rename}>
								<input
									data-department-edit-input
									bind:value={editingName}
									required
									maxlength="60"
									aria-label="Tên phân ban"
									class="h-10 min-w-0 flex-1 rounded-md border-[var(--color-border-strong)]"
								/>
								<button
									type="submit"
									disabled={departmentStore.isSaving || !editingName.trim()}
									class="grid h-10 w-10 place-items-center rounded-md bg-[var(--color-primary)] text-white disabled:opacity-50"
									aria-label="Lưu tên"
									><span class="icon-[lucide--check] h-4 w-4" aria-hidden="true"></span></button
								>
								<button
									type="button"
									class="grid h-10 w-10 place-items-center rounded-md border border-[var(--color-border-strong)]"
									aria-label="Huỷ sửa"
									onclick={() => (editingID = '')}
									><span class="icon-[lucide--x] h-4 w-4" aria-hidden="true"></span></button
								>
							</form>
						{:else}
							<div>
								<div class="min-w-0">
									<div class="flex items-start gap-2">
										<p class="min-w-0 flex-1 text-sm leading-5 font-semibold break-words">
											{item.name}
										</p>
										<span
											class={[
												'shrink-0 rounded px-1.5 py-0.5 text-[10px] font-semibold',
												item.active
													? 'bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]'
													: 'bg-[var(--color-surface-muted)] text-[var(--color-text-muted)]'
											]}>{item.active ? 'Đang dùng' : 'Đã ẩn'}</span
										>
									</div>
									<p class="mt-1 text-xs text-[var(--color-text-secondary)]">
										{item.active_volunteer_count} đang công quả
									</p>
								</div>
								<div
									class="mt-3 flex items-center justify-end gap-1 border-t border-[var(--color-border)] pt-2"
								>
									<button
										type="button"
										class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-primary-dark)]"
										aria-label="Xem danh sách đang công quả"
										title="Xem danh sách đang công quả"
										onclick={() => openVolunteerList(item)}
										><span class="icon-[lucide--list-filter] h-4 w-4" aria-hidden="true"
										></span></button
									>
									<button
										type="button"
										class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-text-secondary)]"
										aria-label="Sửa tên"
										title="Sửa tên"
										onclick={() => beginEdit(item)}
										><span class="icon-[lucide--pencil] h-4 w-4" aria-hidden="true"></span></button
									>
									<button
										type="button"
										class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-text-secondary)]"
										aria-label={item.active ? 'Ngừng sử dụng' : 'Mở lại'}
										title={item.active ? 'Ngừng sử dụng' : 'Mở lại'}
										onclick={() => void toggle(item)}
										><span
											class={item.active
												? 'icon-[lucide--eye-off] h-4 w-4'
												: 'icon-[lucide--eye] h-4 w-4'}
											aria-hidden="true"
										></span></button
									>
									<button
										type="button"
										disabled={item.volunteer_count > 0}
										class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-danger)] disabled:opacity-30"
										aria-label="Xoá phân ban"
										title={item.volunteer_count > 0
											? 'Không thể xoá phân ban đang có Huynh đệ'
											: 'Xoá phân ban'}
										onclick={() => void remove(item)}
										><span class="icon-[lucide--trash-2] h-4 w-4" aria-hidden="true"></span></button
									>
								</div>
							</div>
						{/if}
					</li>
				{/each}
			</ul>
			<div
				class="hidden overflow-hidden rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] md:block"
			>
				<table class="w-full table-fixed border-collapse text-left text-sm">
					<thead
						class="border-b border-[var(--color-border)] bg-[var(--color-surface-muted)] text-xs font-semibold text-[var(--color-text-secondary)]"
					>
						<tr>
							<th class="w-[45%] px-4 py-3">Tên phân ban</th>
							<th class="w-[18%] px-4 py-3">Đang công quả</th>
							<th class="w-[18%] px-4 py-3">Trạng thái</th>
							<th class="w-[19%] px-4 py-3 text-right">Thao tác</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-[var(--color-border)]">
						{#each departmentStore.items as item (item.id)}
							<tr>
								{#if editingID === item.id}
									<td colspan="4" class="px-4 py-3">
										<form class="flex max-w-2xl gap-2" onsubmit={rename}>
											<input
												data-department-edit-input
												bind:value={editingName}
												required
												maxlength="60"
												aria-label="Tên phân ban"
												class="h-10 min-w-0 flex-1 rounded-md border-[var(--color-border-strong)]"
											/>
											<button
												type="submit"
												disabled={departmentStore.isSaving || !editingName.trim()}
												class="grid h-10 w-10 place-items-center rounded-md bg-[var(--color-primary)] text-white disabled:opacity-50"
												aria-label="Lưu tên"
												><span class="icon-[lucide--check] h-4 w-4" aria-hidden="true"
												></span></button
											>
											<button
												type="button"
												class="grid h-10 w-10 place-items-center rounded-md border border-[var(--color-border-strong)]"
												aria-label="Huỷ sửa"
												onclick={() => (editingID = '')}
												><span class="icon-[lucide--x] h-4 w-4" aria-hidden="true"></span></button
											>
										</form>
									</td>
								{:else}
									<td class="truncate px-4 py-3 font-semibold">{item.name}</td>
									<td class="px-4 py-3 text-[var(--color-text-secondary)]"
										>{item.active_volunteer_count} người</td
									>
									<td class="px-4 py-3"
										><span
											class={[
												'inline-flex rounded px-2 py-1 text-xs font-semibold',
												item.active
													? 'bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]'
													: 'bg-[var(--color-surface-muted)] text-[var(--color-text-muted)]'
											]}>{item.active ? 'Đang dùng' : 'Đã ẩn'}</span
										></td
									>
									<td class="px-4 py-2">
										<div class="flex justify-end gap-1">
											<button
												type="button"
												class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-primary-dark)] hover:bg-[var(--color-primary-soft)]"
												aria-label="Xem danh sách đang công quả"
												title="Xem danh sách đang công quả"
												onclick={() => openVolunteerList(item)}
												><span class="icon-[lucide--list-filter] h-4 w-4" aria-hidden="true"
												></span></button
											>
											<button
												type="button"
												class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-muted)]"
												aria-label="Sửa tên"
												title="Sửa tên"
												onclick={() => beginEdit(item)}
												><span class="icon-[lucide--pencil] h-4 w-4" aria-hidden="true"
												></span></button
											>
											<button
												type="button"
												class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-muted)]"
												aria-label={item.active ? 'Ngừng sử dụng' : 'Mở lại'}
												title={item.active ? 'Ngừng sử dụng' : 'Mở lại'}
												onclick={() => void toggle(item)}
												><span
													class={item.active
														? 'icon-[lucide--eye-off] h-4 w-4'
														: 'icon-[lucide--eye] h-4 w-4'}
													aria-hidden="true"
												></span></button
											>
											<button
												type="button"
												disabled={item.volunteer_count > 0}
												class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-danger)] hover:bg-[var(--color-danger-soft)] disabled:opacity-30"
												aria-label="Xoá phân ban"
												title={item.volunteer_count > 0
													? 'Không thể xoá phân ban đang có Huynh đệ'
													: 'Xoá phân ban'}
												onclick={() => void remove(item)}
												><span class="icon-[lucide--trash-2] h-4 w-4" aria-hidden="true"
												></span></button
											>
										</div>
									</td>
								{/if}
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</section>
