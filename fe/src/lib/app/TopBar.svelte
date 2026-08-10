<script lang="ts">
	import { onMount } from 'svelte';
	import { router } from '$lib/navigation/router.svelte';
	import type { AppRoute } from '$lib/navigation/routes';
	import { eventStore } from '$lib/events/event-store.svelte';
	import Logo from '$lib/ui/Logo.svelte';

	let { route }: { route: AppRoute } = $props();
	let menuOpen = $state(false);
	let menuRoot: HTMLDivElement | undefined;
	let showBack = $derived(
		route.name === 'event-detail' || route.name === 'event-new' || route.name === 'event-edit'
	);
	let isFormRoute = $derived(route.name === 'event-new' || route.name === 'event-edit');
	let canSaveForm = $derived(
		Boolean(eventStore.form.title && eventStore.form.starts_at && !eventStore.isSaving)
	);

	function toggleMenu() {
		menuOpen = !menuOpen;
	}

	function closeMenu() {
		menuOpen = false;
	}

	onMount(() => {
		function handlePointerDown(event: PointerEvent) {
			if (!menuOpen || !menuRoot) return;
			if (event.target instanceof Node && menuRoot.contains(event.target)) return;
			closeMenu();
		}

		document.addEventListener('pointerdown', handlePointerDown);

		return () => {
			document.removeEventListener('pointerdown', handlePointerDown);
		};
	});
</script>

<header
	class="z-20 border-b border-[var(--color-border)] bg-[rgb(255_255_255_/_0.94)] px-4 pt-[max(env(safe-area-inset-top),0.75rem)] backdrop-blur"
>
	<div class="flex h-14 items-center justify-between gap-3">
		<div class="flex min-w-0 items-center gap-2">
			{#if showBack}
				<button
					type="button"
					class="grid h-10 w-10 place-items-center rounded-full border border-[var(--color-border-strong)] bg-[var(--color-surface)] text-[var(--color-primary)]"
					aria-label="Quay lại"
					onclick={() => router.back()}
				>
					<span class="icon-[lucide--chevron-left] h-5 w-5" aria-hidden="true"></span>
				</button>
			{/if}
			<div class="flex min-w-0 items-center gap-2">
				<div class="mt-0.5">
					<Logo compact />
				</div>
				<p class="truncate text-lg font-semibold text-[var(--color-text)]">{route.title}</p>
			</div>
		</div>

		<div class="relative shrink-0" bind:this={menuRoot}>
			{#if isFormRoute}
				<button
					type="submit"
					form="event-form"
					class="grid h-10 w-10 place-items-center rounded-full bg-[var(--color-primary)] text-white shadow-[var(--shadow-soft)] disabled:opacity-60"
					disabled={!canSaveForm}
					aria-label={eventStore.isSaving
						? 'Đang lưu'
						: route.name === 'event-edit'
							? 'Lưu'
							: 'Tạo'}
				>
					<span class="icon-[lucide--check] h-5 w-5" aria-hidden="true"></span>
				</button>
			{:else}
				<button
					type="button"
					class="grid h-10 w-10 place-items-center rounded-full bg-[var(--color-primary-soft)] text-sm font-semibold text-[var(--color-primary-dark)]"
					aria-expanded={menuOpen}
					aria-label="Menu tài khoản"
					onclick={toggleMenu}
				>
					<span class="icon-[lucide--user] h-5 w-5" aria-hidden="true"></span>
				</button>
			{/if}

			{#if menuOpen}
				<div
					class="absolute top-12 right-0 w-56 rounded-[12px] border border-[var(--color-border)] bg-[var(--color-surface)] p-2 shadow-[var(--shadow-popover)]"
				>
					<button
						type="button"
						class="flex h-10 w-full items-center gap-2 rounded-[10px] px-3 text-left text-sm text-[var(--color-text)] hover:bg-[var(--color-surface-muted)]"
						onclick={() => {
							closeMenu();
							router.push('/settings');
						}}
					>
						<span class="icon-[lucide--user-round] h-4 w-4" aria-hidden="true"></span>
						Hồ sơ
					</button>
					<button
						type="button"
						class="flex h-10 w-full items-center gap-2 rounded-[10px] px-3 text-left text-sm text-[var(--color-text)] hover:bg-[var(--color-surface-muted)]"
						onclick={() => {
							closeMenu();
							router.push('/settings');
						}}
					>
						<span class="icon-[lucide--bell] h-4 w-4" aria-hidden="true"></span>
						Cài đặt thông báo
					</button>
					<button
						type="button"
						class="flex h-10 w-full items-center gap-2 rounded-[10px] px-3 text-left text-sm text-[var(--color-danger)] hover:bg-[var(--color-danger-soft)]"
						onclick={closeMenu}
					>
						<span class="icon-[lucide--log-out] h-4 w-4" aria-hidden="true"></span>
						Đăng xuất
					</button>
				</div>
			{/if}
		</div>
	</div>
</header>
