<script lang="ts">
	import { eventStore } from '$lib/events/event-store.svelte';
	import { formatEventDate } from '$lib/events/time';
	import { router } from '$lib/navigation/router.svelte';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';

	let { active = false, refreshKey = 0 }: { active?: boolean; refreshKey?: number } = $props();
	let events = $derived(eventStore.orderedEvents());
	const eventsRefreshTTL = 45_000;

	$effect(() => {
		if (!active) return;

		Number(refreshKey);
		void eventStore.refreshIfStale(eventsRefreshTTL);
	});

	function statusLabel(status: string): string {
		if (status === 'active') return 'Đang hoạt động';
		if (status === 'archived') return 'Đã lưu trữ';
		if (status === 'cancelled') return 'Đã huỷ';
		return status;
	}
</script>

<section class="flex h-full flex-col overflow-hidden">
	<div class="border-b border-[var(--color-border)] bg-[var(--color-bg)] px-4 py-3">
		<button
			type="button"
			class="flex h-11 w-full items-center justify-center gap-2 rounded-[12px] bg-[var(--color-primary)] text-sm font-semibold text-white shadow-[var(--shadow-soft)]"
			onclick={() => router.push('/events/new')}
		>
			<span class="icon-[lucide--plus] h-5 w-5" aria-hidden="true"></span>
			Tạo sự kiện
		</button>
	</div>

	<div class="flex-1 overflow-y-auto px-4 py-3">
		{#if eventStore.isLoading}
			<LoadingIndicator label="Đang tải sự kiện..." />
		{:else if events.length === 0}
			<div
				class="rounded-[14px] border border-dashed border-[var(--color-border-strong)] bg-[var(--color-surface)] px-4 py-8 text-center"
			>
				<p class="text-sm font-semibold text-[var(--color-text)]">Chưa có sự kiện</p>
				<p class="mt-1 text-sm text-[var(--color-text-secondary)]">
					Tạo sự kiện đầu tiên để bắt đầu nhắc hẹn.
				</p>
			</div>
		{:else}
			<ul class="space-y-3">
				{#each events as event (event.id)}
					<li>
						<button
							type="button"
							class="w-full rounded-[14px] border border-[var(--color-border)] bg-[var(--color-surface)] p-4 text-left shadow-[var(--shadow-soft)] transition hover:border-[var(--color-border-strong)]"
							onclick={() => router.push(`/events/${encodeURIComponent(event.id)}`)}
						>
							<span class="block truncate text-base font-semibold text-[var(--color-text)]"
								>{event.title}</span
							>
							<span class="mt-2 block text-sm text-[var(--color-text-secondary)]"
								>{formatEventDate(event.starts_at)}</span
							>
							<span
								class="mt-3 inline-flex rounded-[999px] bg-[var(--color-primary-soft)] px-2.5 py-1 text-xs font-medium text-[var(--color-primary-dark)]"
							>
								{statusLabel(event.status)}
							</span>
						</button>
					</li>
				{/each}
			</ul>
			{#if eventStore.hasMore}
				<button
					type="button"
					class="mt-4 flex h-11 w-full items-center justify-center gap-2 rounded-[12px] border border-[var(--color-border-strong)] bg-[var(--color-surface)] text-sm font-semibold text-[var(--color-primary)] disabled:opacity-60"
					disabled={eventStore.isLoadingMore}
					onclick={() => eventStore.loadMore()}
				>
					{#if eventStore.isLoadingMore}
						<LoadingIndicator label="Đang tải thêm..." />
					{:else}
						<span class="icon-[lucide--chevrons-down] h-4 w-4" aria-hidden="true"></span>
						Tải thêm
					{/if}
				</button>
			{/if}
		{/if}
	</div>
</section>
