<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { authStore } from '$lib/auth/auth-store.svelte';
	import { router } from '$lib/navigation/router.svelte';
	import { volunteerStore } from '$lib/volunteers/volunteer-store.svelte';
	import LoginScreen from '$lib/screens/LoginScreen.svelte';
	import VolunteerListScreen from '$lib/screens/VolunteerListScreen.svelte';
	import VolunteerFormScreen from '$lib/screens/VolunteerFormScreen.svelte';
	import VolunteerDetailScreen from '$lib/screens/VolunteerDetailScreen.svelte';
	import UsersScreen from '$lib/screens/UsersScreen.svelte';
	import DepartmentsScreen from '$lib/screens/DepartmentsScreen.svelte';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';
	import Popup from '$lib/ui/Popup.svelte';
	import { popupStore } from '$lib/ui/popup-store.svelte';
	import Toast from '$lib/ui/Toast.svelte';
	import { toastStore } from '$lib/ui/toast-store.svelte';
	import BottomNav from './BottomNav.svelte';
	import TopBar from './TopBar.svelte';

	let route = $derived(router.current);
	let isFormRoute = $derived(route.name === 'volunteer-new' || route.name === 'volunteer-edit');

	onMount(async () => {
		router.init();
		await authStore.init();
	});

	onDestroy(() => router.destroy());

	$effect(() => {
		if (route.name === 'volunteer-new') volunteerStore.prepareCreate();
		if (route.name === 'volunteer-edit' && route.volunteerId)
			void volunteerStore.prepareEdit(route.volunteerId);
	});
</script>

{#if authStore.initializing}
	<div class="grid min-h-screen place-items-center">
		<LoadingIndicator label="Đang khởi động..." />
	</div>
{:else if !authStore.user}
	<LoginScreen />
{:else}
	<div
		class="mx-auto flex min-h-screen max-w-lg flex-col bg-[var(--color-bg)] text-[var(--color-text)] shadow-[0_0_0_1px_var(--color-border)]"
	>
		<TopBar {route} />
		<section class="relative flex-1 overflow-hidden">
			{#if route.name === 'volunteers'}
				<VolunteerListScreen />
			{:else if route.name === 'volunteer-detail' && route.volunteerId}
				{#key route.volunteerId}<VolunteerDetailScreen volunteerId={route.volunteerId} />{/key}
			{:else if route.name === 'volunteer-new' || route.name === 'volunteer-edit'}
				{#key route.path}<VolunteerFormScreen volunteerId={route.volunteerId} />{/key}
			{:else if route.name === 'users'}
				<UsersScreen />
			{:else if route.name === 'departments'}
				<DepartmentsScreen />
			{/if}
		</section>
		{#if !isFormRoute}<BottomNav {route} />{/if}
	</div>
{/if}

<Toast
	open={toastStore.open}
	message={toastStore.message}
	tone={toastStore.tone}
	onClose={() => toastStore.close()}
/>
<Popup open={popupStore.open} title={popupStore.title} onClose={() => popupStore.cancel()}>
	<p class="text-sm leading-6 text-[var(--color-text-secondary)]">{popupStore.message}</p>
	{#snippet footer()}
		<div class="grid grid-cols-2 gap-3">
			<button
				type="button"
				class="h-11 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] text-sm font-semibold"
				onclick={() => popupStore.cancel()}>{popupStore.cancelLabel}</button
			>
			<button
				type="button"
				class={[
					'h-11 rounded-md text-sm font-semibold text-white',
					popupStore.tone === 'danger' ? 'bg-[var(--color-danger)]' : 'bg-[var(--color-primary)]'
				]}
				onclick={() => popupStore.accept()}>{popupStore.confirmLabel}</button
			>
		</div>
	{/snippet}
</Popup>
