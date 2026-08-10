<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { AppRoute } from '$lib/navigation/routes';
	import { router } from '$lib/navigation/router.svelte';
	import { volunteerStore } from '$lib/volunteers/volunteer-store.svelte';

	let { route, children }: { route: AppRoute; children: Snippet } = $props();
	let isForm = $derived(route.name === 'volunteer-new' || route.name === 'volunteer-edit');
	let canSave = $derived(
		Boolean(
			volunteerStore.form.full_name && volunteerStore.form.arrival_date && !volunteerStore.isSaving
		)
	);

	function close() {
		router.push('/volunteers');
	}
</script>

<svelte:window onkeydown={(event) => event.key === 'Escape' && close()} />

<div class="h-full md:absolute md:inset-0 md:z-30 md:flex md:items-center md:justify-center md:p-6">
	<button
		type="button"
		class="absolute inset-0 hidden h-full w-full cursor-default bg-[rgb(24_32_28_/_0.42)] md:block"
		aria-label="Đóng popup"
		onclick={close}
	></button>
	<div
		class={[
			'relative z-10 flex h-full w-full flex-col bg-[var(--color-bg)] md:max-h-[calc(100dvh-8rem)] md:overflow-hidden md:rounded-md md:border md:border-[var(--color-border)] md:bg-[var(--color-surface)] md:shadow-[var(--shadow-popover)]',
			isForm ? 'md:max-w-5xl' : 'md:max-w-4xl'
		]}
		role="dialog"
		aria-modal="true"
		aria-label={route.title}
	>
		<header
			class="hidden h-16 shrink-0 items-center justify-between gap-4 border-b border-[var(--color-border)] px-5 md:flex"
		>
			<div class="min-w-0">
				<p class="truncate text-base font-semibold">{route.title}</p>
				<p class="mt-0.5 text-xs text-[var(--color-text-secondary)]">Hồ sơ Huynh đệ công quả</p>
			</div>
			<div class="flex items-center gap-2">
				{#if isForm}
					<button
						type="submit"
						form="volunteer-form"
						disabled={!canSave}
						class="flex h-10 items-center justify-center gap-2 rounded-md bg-[var(--color-primary)] px-4 text-sm font-semibold text-white disabled:opacity-50"
					>
						<span class="icon-[lucide--check] h-4 w-4" aria-hidden="true"></span>
						Lưu hồ sơ
					</button>
				{/if}
				<button
					type="button"
					class="grid h-10 w-10 place-items-center rounded-md border border-[var(--color-border-strong)] text-[var(--color-text-secondary)]"
					aria-label="Đóng"
					title="Đóng"
					onclick={close}
				>
					<span class="icon-[lucide--x] h-5 w-5" aria-hidden="true"></span>
				</button>
			</div>
		</header>
		<div class="min-h-0 flex-1">{@render children()}</div>
	</div>
</div>
