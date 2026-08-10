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
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';

	let searchTimer: ReturnType<typeof setTimeout> | undefined;
	const searchDebounceMs = 350;
	const listCacheTtlMs = 45_000;

	function search() {
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => void volunteerStore.load(), searchDebounceMs);
	}

	onMount(() => {
		let currentDate = vietnamDateKey();
		volunteerStore.status = 'active';
		void volunteerStore.refreshIfStale(listCacheTtlMs);

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
	<div class="space-y-3 border-b border-[var(--color-border)] bg-[var(--color-bg)] px-4 py-3">
		<button
			type="button"
			class="flex h-11 w-full items-center justify-center gap-2 rounded-md bg-[var(--color-primary)] text-sm font-semibold text-white"
			onclick={() => router.push('/volunteers/new')}
		>
			<span class="icon-[lucide--user-plus] h-5 w-5" aria-hidden="true"></span>
			Thêm Huynh đệ công quả
		</button>
		<div class="flex gap-2">
			<label class="relative min-w-0 flex-1">
				<span
					class="absolute top-3 left-3 icon-[lucide--search] h-4 w-4 text-[var(--color-text-muted)]"
					aria-hidden="true"
				></span>
				<input
					bind:value={volunteerStore.query}
					oninput={search}
					placeholder="Tìm theo tên, pháp danh, SĐT"
					aria-label="Tìm hồ sơ"
					class="h-10 w-full rounded-md border-[var(--color-border-strong)] pr-3 pl-9 text-sm"
				/>
			</label>
			<select
				bind:value={volunteerStore.status}
				onchange={() => volunteerStore.load()}
				aria-label="Lọc trạng thái"
				class="h-10 rounded-md border-[var(--color-border-strong)] pr-8 text-sm"
			>
				<option value="active">Đang công quả</option>
				<option value="departed">Đã ra về</option>
				<option value="">Tất cả</option>
			</select>
		</div>
	</div>

	<div class="flex-1 overflow-y-auto px-4 py-3">
		{#if volunteerStore.isLoading}
			<LoadingIndicator label="Đang tải hồ sơ..." />
		{:else if volunteerStore.items.length === 0}
			<div
				class="border border-dashed border-[var(--color-border-strong)] bg-[var(--color-surface)] px-4 py-10 text-center"
			>
				<span
					class="mx-auto icon-[lucide--users] block h-8 w-8 text-[var(--color-text-muted)]"
					aria-hidden="true"
				></span>
				<p class="mt-3 text-sm font-semibold">Chưa có hồ sơ phù hợp</p>
			</div>
		{:else}
			<ul class="space-y-2">
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
									<span class="mt-1 block truncate text-xs font-medium text-[var(--color-primary)]">
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
		{/if}
	</div>
</section>
