<script lang="ts">
	let { label = 'Đang tải...' }: { label?: string } = $props();
</script>

<div
	class="inline-flex items-center gap-2 text-sm text-[var(--color-text-secondary)]"
	role="status"
	aria-live="polite"
>
	<span class="icon-[lucide--loader-circle] h-4 w-4 animate-spin" aria-hidden="true"></span>
	<span>{label}</span>
</div>
