<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		open = false,
		title = '',
		onClose,
		children,
		footer
	}: {
		open?: boolean;
		title?: string;
		onClose?: () => void;
		children?: Snippet;
		footer?: Snippet;
	} = $props();
</script>

{#if open}
	<div class="fixed inset-0 z-40 grid place-items-center px-4" role="presentation">
		<button
			type="button"
			class="absolute inset-0 h-full w-full cursor-default bg-[rgb(24_32_28_/_0.34)]"
			aria-label="Đóng popup"
			onclick={onClose}
		></button>
		<div
			class="relative z-10 w-full max-w-sm rounded-[14px] border border-[var(--color-border)] bg-[var(--color-surface)] p-4 shadow-[var(--shadow-popover)]"
			role="dialog"
			aria-modal="true"
			aria-label={title}
			tabindex="-1"
		>
			<div class="flex items-center justify-between gap-3">
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
			<div class="mt-3">
				{@render children?.()}
			</div>
			{#if footer}
				<div class="mt-4">
					{@render footer()}
				</div>
			{/if}
		</div>
	</div>
{/if}
