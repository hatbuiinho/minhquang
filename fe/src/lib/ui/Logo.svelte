<script lang="ts">
	let { compact = false }: { compact?: boolean } = $props();
</script>

<div class="inline-flex items-center gap-2 text-[var(--color-text)]" aria-label="Nhắc việc">
	<span class="grid h-9 w-9 place-items-center rounded-[10px] bg-[var(--color-primary)] text-white">
		<span class="icon-[lucide--bell-ring] h-5 w-5" aria-hidden="true"></span>
	</span>
	{#if !compact}
		<span class="text-base font-semibold">Nhắc việc</span>
	{/if}
</div>
