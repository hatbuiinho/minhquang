<script lang="ts">
	type ToastTone = 'success' | 'error' | 'info';

	let {
		message = '',
		tone = 'info',
		open = false,
		onClose
	}: {
		message?: string;
		tone?: ToastTone;
		open?: boolean;
		onClose?: () => void;
	} = $props();

	let toneClass = $derived(
		tone === 'success'
			? 'border-[var(--color-primary)] bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]'
			: tone === 'error'
				? 'border-[var(--color-danger)] bg-[var(--color-danger-soft)] text-[var(--color-danger)]'
				: 'border-[var(--color-border-strong)] bg-[var(--color-surface)] text-[var(--color-text-secondary)]'
	);
</script>

{#if open && message}
	<div
		class="fixed right-4 bottom-24 left-4 z-50 md:top-5 md:right-5 md:bottom-auto md:left-auto md:w-full md:max-w-sm"
	>
		<div
			class={[
				'flex min-h-12 items-center gap-3 rounded-[12px] border px-3 py-2 shadow-[var(--shadow-popover)]',
				toneClass
			]}
		>
			<span
				class={tone === 'success'
					? 'icon-[lucide--circle-check] h-5 w-5 shrink-0'
					: tone === 'error'
						? 'icon-[lucide--circle-alert] h-5 w-5 shrink-0'
						: 'icon-[lucide--info] h-5 w-5 shrink-0'}
				aria-hidden="true"
			></span>
			<p class="min-w-0 flex-1 text-sm leading-5 break-words">{message}</p>
			<button
				type="button"
				class="grid h-8 w-8 shrink-0 place-items-center"
				aria-label="Đóng"
				onclick={onClose}
			>
				<span class="icon-[lucide--x] h-4 w-4" aria-hidden="true"></span>
			</button>
		</div>
	</div>
{/if}
