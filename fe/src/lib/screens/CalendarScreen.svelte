<script lang="ts">
	import { eventStore } from '$lib/events/event-store.svelte';

	let { active = false, refreshKey = 0 }: { active?: boolean; refreshKey?: number } = $props();
	const eventsRefreshTTL = 45_000;

	$effect(() => {
		if (!active) return;

		Number(refreshKey);
		void eventStore.refreshIfStale(eventsRefreshTTL);
	});
</script>

<section class="h-full overflow-y-auto px-4 py-4">
	<div
		class="rounded-[16px] border border-[var(--color-border)] bg-[var(--color-surface)] p-5 shadow-[var(--shadow-soft)]"
	>
		<h2 class="text-lg font-semibold text-[var(--color-text)]">Lịch</h2>
		<p class="mt-2 text-sm leading-6 text-[var(--color-text-secondary)]">
			Màn lịch sẽ dùng chung dữ liệu sự kiện.
		</p>
	</div>
</section>
