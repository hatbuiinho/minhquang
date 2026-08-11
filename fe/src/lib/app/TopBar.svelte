<script lang="ts">
	import type { AppRoute } from '$lib/navigation/routes';
	import { router } from '$lib/navigation/router.svelte';
	import { volunteerStore } from '$lib/volunteers/volunteer-store.svelte';
	import { authStore } from '$lib/auth/auth-store.svelte';
	import Logo from '$lib/ui/Logo.svelte';

	let { route }: { route: AppRoute } = $props();
	let showBack = $derived(route.name.startsWith('volunteer-'));
	let desktopTitle = $derived(showBack ? 'Huynh đệ công quả' : route.title);
	let isForm = $derived(route.name === 'volunteer-new' || route.name === 'volunteer-edit');
	let canSave = $derived(
		Boolean(
			volunteerStore.form.full_name && volunteerStore.form.arrival_date && !volunteerStore.isSaving
		)
	);
</script>

<header
	class="z-20 border-b border-[var(--color-border)] bg-[rgb(255_255_255_/_0.96)] px-4 pt-[max(env(safe-area-inset-top),0.75rem)] backdrop-blur md:px-6 md:pt-0 lg:px-8"
>
	<div class="flex h-14 items-center justify-between gap-3 md:h-[72px]">
		<div class="flex min-w-0 items-center gap-2">
			{#if showBack}
				<button
					type="button"
					class="grid h-10 w-10 place-items-center rounded-full border border-[var(--color-border-strong)] md:hidden"
					aria-label="Quay lại"
					onclick={() => router.back()}
					><span class="icon-[lucide--chevron-left] h-5 w-5" aria-hidden="true"></span></button
				>
			{/if}
			<span class="md:hidden"><Logo compact /></span>
			<p class="truncate text-base font-semibold md:text-lg">
				<span class="md:hidden">{route.title}</span>
				<span class="hidden md:inline">{desktopTitle}</span>
			</p>
		</div>
		{#if isForm}
			<button
				type="submit"
				form="volunteer-form"
				disabled={!canSave}
				class="flex h-10 w-10 items-center justify-center rounded-md bg-[var(--color-primary)] text-white disabled:opacity-50 md:hidden"
				aria-label="Lưu Huynh đệ"
				><span class="icon-[lucide--check] h-5 w-5" aria-hidden="true"></span></button
			>
		{:else}
			<button
				type="button"
				class="grid h-10 w-10 place-items-center rounded-full bg-[var(--color-surface-muted)] text-[var(--color-text-secondary)] md:hidden"
				aria-label="Đăng xuất"
				title="Đăng xuất"
				onclick={() => authStore.logout()}
				><span class="icon-[lucide--log-out] h-5 w-5" aria-hidden="true"></span></button
			>
		{/if}
	</div>
</header>
