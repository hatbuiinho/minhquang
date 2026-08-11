<script module lang="ts">
	function formatDate(value: string) {
		return new Intl.DateTimeFormat('vi-VN').format(new Date(value));
	}
</script>

<script lang="ts">
	import { onMount } from 'svelte';
	import { router } from '$lib/navigation/router.svelte';
	import { volunteerStore } from '$lib/volunteers/volunteer-store.svelte';
	import { vietnamDateKey } from '$lib/volunteers/status';
	import type { VolunteerSortKey } from '$lib/volunteers/api';
	import { listDepartments, type Department } from '$lib/departments/api';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';

	let searchTimer: ReturnType<typeof setTimeout> | undefined;
	let departments = $state<Department[]>([]);
	const searchDebounceMs = 350;
	const listCacheTtlMs = 45_000;

	function search() {
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => void volunteerStore.load(), searchDebounceMs);
	}

	function toggleSort(key: VolunteerSortKey) {
		if (volunteerStore.sortKey === key) {
			volunteerStore.sortDirection = volunteerStore.sortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			volunteerStore.sortKey = key;
			volunteerStore.sortDirection =
				key === 'arrival_date' || key === 'departure_date' ? 'desc' : 'asc';
		}
		void volunteerStore.loadSorted();
	}

	function changeDepartment() {
		volunteerStore.departmentName =
			departments.find((item) => item.id === volunteerStore.departmentId)?.name ?? '';
		void volunteerStore.load();
	}

	onMount(() => {
		let currentDate = vietnamDateKey();
		volunteerStore.status = 'active';
		void volunteerStore.refreshIfStale(listCacheTtlMs);
		void listDepartments('', 'all')
			.then((items) => (departments = items))
			.catch(() => (departments = []));

		function smartRefresh() {
			currentDate = vietnamDateKey();
			void volunteerStore.refreshIfStale(listCacheTtlMs);
		}

		function refreshWhenVisible() {
			if (document.visibilityState === 'visible') smartRefresh();
		}

		const dateTimer = window.setInterval(() => {
			const nextDate = vietnamDateKey();
			if (nextDate === currentDate) return;
			currentDate = nextDate;
			void volunteerStore.load();
		}, 60_000);

		window.addEventListener('focus', smartRefresh);
		document.addEventListener('visibilitychange', refreshWhenVisible);
		return () => {
			if (searchTimer) clearTimeout(searchTimer);
			window.clearInterval(dateTimer);
			window.removeEventListener('focus', smartRefresh);
			document.removeEventListener('visibilitychange', refreshWhenVisible);
		};
	});
</script>

<section class="flex h-full flex-col overflow-hidden">
	<div
		class="border-b border-[var(--color-border)] bg-[var(--color-bg)] px-4 py-3 md:px-6 md:py-4 lg:px-8"
	>
		<div class="mx-auto flex max-w-[1320px] flex-col gap-3 md:flex-row md:items-center">
			<button
				type="button"
				class="hidden h-11 items-center justify-center gap-2 rounded-md bg-[var(--color-primary)] px-4 text-sm font-semibold text-white md:order-2 md:flex md:w-auto"
				onclick={() => router.push('/volunteers/new')}
			>
				<span class="icon-[lucide--user-plus] h-5 w-5" aria-hidden="true"></span>
				Thêm Huynh đệ công quả
			</button>
			<div class="grid min-w-0 flex-1 grid-cols-2 gap-2 md:order-1 md:flex">
				<label class="relative col-span-2 min-w-0 md:flex-1">
					<span
						class="absolute top-3 left-3 icon-[lucide--search] h-4 w-4 text-[var(--color-text-muted)]"
						aria-hidden="true"
					></span>
					<input
						bind:value={volunteerStore.query}
						oninput={search}
						placeholder="Tìm theo tên, pháp danh, SĐT"
						aria-label="Tìm Huynh đệ"
						class="h-10 w-full rounded-md border-[var(--color-border-strong)] pr-3 pl-9 text-sm"
					/>
				</label>
				<select
					bind:value={volunteerStore.departmentId}
					onchange={changeDepartment}
					aria-label="Lọc phân ban"
					class="h-10 min-w-0 rounded-md border-[var(--color-border-strong)] pr-8 text-sm md:max-w-52"
				>
					<option value="">Tất cả phân ban</option>
					{#if volunteerStore.departmentId && !departments.some((item) => item.id === volunteerStore.departmentId)}
						<option value={volunteerStore.departmentId}>{volunteerStore.departmentName}</option>
					{/if}
					{#each departments as department (department.id)}
						<option value={department.id}>{department.name}</option>
					{/each}
				</select>
				<select
					bind:value={volunteerStore.status}
					onchange={() => volunteerStore.load()}
					aria-label="Lọc trạng thái"
					class="h-10 min-w-0 rounded-md border-[var(--color-border-strong)] pr-8 text-sm"
				>
					<option value="active">Đang công quả</option>
					<option value="departed">Đã ra về</option>
					<option value="">Tất cả</option>
				</select>
			</div>
		</div>
	</div>

	<div class="flex-1 overflow-y-auto px-4 pt-3 pb-20 md:px-6 md:py-5 lg:px-8">
		<div class="mx-auto max-w-[1320px]">
			<div class="mb-3 flex min-h-5 items-center gap-2 text-xs text-[var(--color-text-secondary)]">
				<span class="icon-[lucide--users] h-4 w-4 shrink-0" aria-hidden="true"></span>
				<p aria-live="polite">
					{volunteerStore.isLoading
						? 'Đang cập nhật số lượng...'
						: `${volunteerStore.total} Huynh đệ`}
					{#if !volunteerStore.isLoading && volunteerStore.departmentName}
						<span> · {volunteerStore.departmentName}</span>
					{/if}
				</p>
				{#if volunteerStore.isRefreshing}
					<span
						class="ml-0.5 icon-[lucide--loader-circle] h-3.5 w-3.5 animate-spin text-[var(--color-primary)]"
						aria-label="Đang cập nhật danh sách"
					></span>
				{/if}
			</div>
			{#if volunteerStore.isLoading}
				<div class="py-16"><LoadingIndicator label="Đang tải Huynh đệ..." /></div>
			{:else if volunteerStore.items.length === 0}
				<div
					class="border border-dashed border-[var(--color-border-strong)] bg-[var(--color-surface)] px-4 py-10 text-center"
				>
					<span
						class="mx-auto icon-[lucide--users] block h-8 w-8 text-[var(--color-text-muted)]"
						aria-hidden="true"
					></span>
					<p class="mt-3 text-sm font-semibold">Chưa có Huynh đệ phù hợp</p>
				</div>
			{:else}
				<ul class="space-y-2 lg:hidden">
					{#each volunteerStore.items as item (item.id)}
						<li>
							<button
								type="button"
								class="flex w-full items-center gap-3 rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] p-3 text-left"
								onclick={() => router.push(`/volunteers/${encodeURIComponent(item.id)}`)}
							>
								{#if item.avatar_url}
									<img
										src={item.avatar_url}
										alt=""
										class="h-12 w-12 shrink-0 rounded-full object-cover"
									/>
								{:else}
									<span
										class="grid h-12 w-12 shrink-0 place-items-center rounded-full bg-[var(--color-primary-soft)] font-semibold text-[var(--color-primary-dark)]"
										>{item.full_name.slice(0, 1).toUpperCase()}</span
									>
								{/if}
								<span class="min-w-0 flex-1">
									<span class="block truncate font-semibold">{item.full_name}</span>
									<span class="mt-0.5 block truncate text-sm text-[var(--color-text-secondary)]"
										>{item.dharma_name || 'Chưa có pháp danh'}{item.cultivation_place
											? ` · ${item.cultivation_place}`
											: ''}</span
									>
									<span class="mt-1 block text-xs leading-5 text-[var(--color-text-muted)]">
										Đến {formatDate(item.arrival_date)} · {item.departure_date
											? `Về ${formatDate(item.departure_date)}`
											: 'Chưa có ngày về'}
									</span>
									{#if item.department}
										<span
											class="mt-1 block truncate text-xs font-medium text-[var(--color-primary)]"
										>
											{item.department}
										</span>
									{/if}
								</span>
								<span
									class="icon-[lucide--chevron-right] h-5 w-5 text-[var(--color-text-muted)]"
									aria-hidden="true"
								></span>
							</button>
						</li>
					{/each}
				</ul>
				<div
					class="hidden max-h-[calc(100dvh-10rem)] overflow-auto rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] lg:block"
					aria-busy={volunteerStore.isRefreshing}
				>
					<table class="w-full min-w-[1320px] border-collapse text-left text-sm">
						<thead
							class="sticky top-0 z-10 border-b border-[var(--color-border)] bg-[var(--color-surface-muted)] text-xs font-semibold text-[var(--color-text-secondary)] shadow-[0_1px_0_var(--color-border)]"
						>
							<tr>
								<th class="w-52 px-4 py-3">{@render sortHeader('full_name', 'Họ tên')}</th>
								<th class="w-40 px-4 py-3">{@render sortHeader('dharma_name', 'Pháp danh')}</th>
								<th class="w-36 px-4 py-3">{@render sortHeader('birth_date', 'Ngày sinh')}</th>
								<th class="w-48 px-4 py-3"
									>{@render sortHeader('cultivation_place', 'Nơi sinh hoạt')}</th
								>
								<th class="w-40 px-4 py-3">{@render sortHeader('department', 'Phân ban')}</th>
								<th class="w-36 px-4 py-3">{@render sortHeader('phone', 'Điện thoại')}</th>
								<th class="w-36 px-4 py-3">{@render sortHeader('arrival_date', 'Ngày đến')}</th>
								<th class="w-36 px-4 py-3">{@render sortHeader('departure_date', 'Ngày về')}</th>
								<th class="w-36 px-4 py-3">{@render sortHeader('status', 'Trạng thái')}</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-[var(--color-border)]">
							{#each volunteerStore.items as item (item.id)}
								<tr
									class="cursor-pointer hover:bg-[var(--color-surface-muted)]"
									onclick={() => router.push(`/volunteers/${encodeURIComponent(item.id)}`)}
								>
									<td class="px-4 py-3">
										<div class="flex min-w-0 items-center gap-3">
											{#if item.avatar_url}
												<img
													src={item.avatar_url}
													alt=""
													class="h-9 w-9 shrink-0 rounded-full object-cover"
												/>
											{:else}
												<span
													class="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-[var(--color-primary-soft)] font-semibold text-[var(--color-primary-dark)]"
													>{item.full_name.slice(0, 1).toUpperCase()}</span
												>
											{/if}
											<div class="min-w-0">
												<p class="truncate font-semibold">{item.full_name}</p>
											</div>
										</div>
									</td>
									<td class="max-w-40 truncate px-4 py-3 text-[var(--color-text-secondary)]"
										>{item.dharma_name || 'Chưa cập nhật'}</td
									>
									<td class="truncate px-4 py-3 text-[var(--color-text-secondary)]"
										>{item.birth_date || 'Chưa cập nhật'}</td
									>
									<td class="max-w-48 truncate px-4 py-3"
										>{item.cultivation_place || 'Chưa cập nhật'}</td
									>
									<td class="truncate px-4 py-3 text-[var(--color-primary-dark)]"
										>{item.department || 'Chưa phân ban'}</td
									>
									<td class="truncate px-4 py-3">{item.phone || 'Chưa cập nhật'}</td>
									<td class="px-4 py-3 whitespace-nowrap">{formatDate(item.arrival_date)}</td>
									<td class="px-4 py-3 whitespace-nowrap text-[var(--color-text-secondary)]"
										>{item.departure_date ? formatDate(item.departure_date) : 'Chưa xác định'}</td
									>
									<td class="px-4 py-3">
										<span
											class={[
												'inline-flex rounded px-2 py-1 text-xs font-semibold whitespace-nowrap',
												item.status === 'departed'
													? 'bg-[var(--color-surface-muted)] text-[var(--color-text-secondary)]'
													: 'bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]'
											]}>{item.status === 'departed' ? 'Đã ra về' : 'Đang công quả'}</span
										>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
			{#if !volunteerStore.isLoading && volunteerStore.hasMore}
				<div class="mt-4 flex flex-col items-center gap-1.5">
					<button
						type="button"
						disabled={volunteerStore.isLoadingMore}
						class="flex h-10 items-center justify-center gap-2 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-5 text-sm font-semibold text-[var(--color-primary-dark)] disabled:opacity-60"
						onclick={() => void volunteerStore.loadMore()}
					>
						<span
							class={volunteerStore.isLoadingMore
								? 'icon-[lucide--loader-circle] h-4 w-4 animate-spin'
								: 'icon-[lucide--chevron-down] h-4 w-4'}
							aria-hidden="true"
						></span>
						{volunteerStore.isLoadingMore ? 'Đang tải...' : 'Xem thêm'}
					</button>
					<p class="text-xs text-[var(--color-text-muted)]">
						Đã hiển thị {volunteerStore.items.length}/{volunteerStore.total}
					</p>
				</div>
			{/if}
		</div>
	</div>

	<button
		type="button"
		class="absolute right-4 bottom-4 z-10 grid h-14 w-14 place-items-center rounded-full bg-[var(--color-primary)] text-white shadow-[var(--shadow-popover)] md:hidden"
		aria-label="Thêm Huynh đệ công quả"
		title="Thêm Huynh đệ công quả"
		onclick={() => router.push('/volunteers/new')}
	>
		<span class="icon-[lucide--user-plus] h-6 w-6" aria-hidden="true"></span>
	</button>
</section>

{#snippet sortHeader(key: VolunteerSortKey, label: string)}
	<button
		type="button"
		class="flex w-full items-center gap-1.5 text-left hover:text-[var(--color-text)]"
		aria-busy={volunteerStore.isSorting && volunteerStore.sortKey === key}
		onclick={() => toggleSort(key)}
	>
		<span>{label}</span>
		<span
			class={[
				'h-3.5 w-3.5 shrink-0',
				volunteerStore.isSorting && volunteerStore.sortKey === key
					? 'icon-[lucide--loader-circle] animate-spin text-[var(--color-primary)]'
					: volunteerStore.sortKey === key
						? volunteerStore.sortDirection === 'asc'
							? 'icon-[lucide--arrow-up] text-[var(--color-primary)]'
							: 'icon-[lucide--arrow-down] text-[var(--color-primary)]'
						: 'icon-[lucide--arrow-up-down] text-[var(--color-text-muted)]'
			]}
			aria-hidden="true"
		></span>
	</button>
{/snippet}
