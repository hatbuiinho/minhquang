<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		open = false,
		title = '',
		onClose,
		children
	}: {
		open?: boolean;
		title?: string;
		onClose?: () => void;
		children?: Snippet;
	} = $props();
</script>

{#if open}
	<div
		class="fixed inset-0 z-40 bg-[rgb(24_32_28_/_0.34)]"
		role="presentation"
		onclick={onClose}
	></div>
	<div
		class="fixed right-0 bottom-0 left-0 z-50 rounded-t-[18px] border-t border-[var(--color-border)] bg-[var(--color-surface)] px-4 pt-3 pb-[max(env(safe-area-inset-bottom),1rem)] shadow-[var(--shadow-popover)]"
		role="dialog"
		aria-modal="true"
		aria-label={title}
	>
		<div class="mb-3 flex items-center justify-between gap-3">
			<h2 class="truncate text-base font-semibold text-[var(--color-text)]">{title}</h2>
			<button
				type="button"
				class="grid h-9 w-9 place-items-center rounded-full text-[var(--color-text-muted)] hover:bg-[var(--color-surface-muted)] hover:text-[var(--color-text)]"
				aria-label="Đóng"
				onclick={onClose}
			>
				<span class="icon-[lucide--x] h-5 w-5" aria-hidden="true"></span>
			</button>
		</div>
		{@render children?.()}
	</div>
{/if}
