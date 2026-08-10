<script lang="ts">
	import { onDestroy, onMount } from 'svelte';

	import { eventStore } from '$lib/events/event-store.svelte';
	import { router } from '$lib/navigation/router.svelte';
	import Popup from '$lib/ui/Popup.svelte';
	import { popupStore } from '$lib/ui/popup-store.svelte';
	import Toast from '$lib/ui/Toast.svelte';
	import { toastStore } from '$lib/ui/toast-store.svelte';
	import { otaUpdateStore } from '$lib/updates/ota-update-store.svelte';
	import BottomNav from './BottomNav.svelte';
	import TopBar from './TopBar.svelte';
	import CalendarScreen from '$lib/screens/CalendarScreen.svelte';
	import EventDetailScreen from '$lib/screens/EventDetailScreen.svelte';
	import EventFormScreen from '$lib/screens/EventFormScreen.svelte';
	import EventListScreen from '$lib/screens/EventListScreen.svelte';
	import RemindersScreen from '$lib/screens/RemindersScreen.svelte';
	import SettingsScreen from '$lib/screens/SettingsScreen.svelte';

	let route = $derived(router.current);
	let isFormRoute = $derived(route.name === 'event-new' || route.name === 'event-edit');
	let refreshKey = $state(0);
	let nextRefreshKey = 0;

	function requestActiveRefresh() {
		refreshKey = ++nextRefreshKey;
	}

	onMount(() => {
		router.init();
		void eventStore.loadOnce();
		void otaUpdateStore.checkOnce();

		function handleVisibilityChange() {
			if (document.visibilityState === 'visible') {
				requestActiveRefresh();
			}
		}

		window.addEventListener('focus', requestActiveRefresh);
		document.addEventListener('visibilitychange', handleVisibilityChange);

		return () => {
			window.removeEventListener('focus', requestActiveRefresh);
			document.removeEventListener('visibilitychange', handleVisibilityChange);
		};
	});

	onDestroy(() => {
		router.destroy();
	});

	$effect(() => {
		const activePath = route.path;
		if (!activePath) return;

		requestActiveRefresh();
	});

	$effect(() => {
		if (route.name === 'event-new') {
			eventStore.prepareCreate();
			return;
		}

		if (route.name === 'event-edit' && route.eventId) {
			eventStore.prepareEdit(route.eventId);
		}
	});
</script>

<div
	class="mx-auto flex min-h-screen max-w-md flex-col bg-[var(--color-bg)] text-[var(--color-text)] shadow-[0_0_0_1px_var(--color-border)] sm:max-w-lg"
>
	<TopBar {route} />

	<section class="relative flex-1 overflow-hidden">
		<div
			class="absolute inset-0"
			class:hidden={route.name !== 'events'}
			aria-hidden={route.name !== 'events'}
		>
			<EventListScreen active={route.name === 'events'} {refreshKey} />
		</div>
		<div
			class="absolute inset-0"
			class:hidden={route.name !== 'event-detail'}
			aria-hidden={route.name !== 'event-detail'}
		>
			{#if route.name === 'event-detail' && route.eventId}
				{#key route.eventId}
					<EventDetailScreen eventId={route.eventId} />
				{/key}
			{/if}
		</div>
		<div
			class="absolute inset-0"
			class:hidden={route.name !== 'event-new' && route.name !== 'event-edit'}
			aria-hidden={route.name !== 'event-new' && route.name !== 'event-edit'}
		>
			{#if route.name === 'event-new' || route.name === 'event-edit'}
				{#key route.path}
					<EventFormScreen
						mode={route.name === 'event-edit' ? 'edit' : 'create'}
						eventId={route.eventId}
					/>
				{/key}
			{/if}
		</div>
		<div
			class="absolute inset-0"
			class:hidden={route.name !== 'calendar'}
			aria-hidden={route.name !== 'calendar'}
		>
			<CalendarScreen active={route.name === 'calendar'} {refreshKey} />
		</div>
		<div
			class="absolute inset-0"
			class:hidden={route.name !== 'reminders'}
			aria-hidden={route.name !== 'reminders'}
		>
			<RemindersScreen active={route.name === 'reminders'} {refreshKey} />
		</div>
		<div
			class="absolute inset-0"
			class:hidden={route.name !== 'settings'}
			aria-hidden={route.name !== 'settings'}
		>
			<SettingsScreen />
		</div>
	</section>

	{#if !isFormRoute}
		<BottomNav {route} />
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
					class="h-11 rounded-[10px] border border-[var(--color-border-strong)] bg-[var(--color-surface)] text-sm font-semibold text-[var(--color-text)]"
					onclick={() => popupStore.cancel()}
				>
					{popupStore.cancelLabel}
				</button>
				<button
					type="button"
					class={[
						'flex h-11 items-center justify-center gap-2 rounded-[10px] text-sm font-semibold text-white',
						popupStore.tone === 'danger' ? 'bg-[var(--color-danger)]' : 'bg-[var(--color-primary)]'
					]}
					onclick={() => popupStore.accept()}
				>
					<span
						class={popupStore.tone === 'danger'
							? 'icon-[lucide--trash-2] h-4 w-4'
							: 'icon-[lucide--check] h-4 w-4'}
						aria-hidden="true"
					></span>
					{popupStore.confirmLabel}
				</button>
			</div>
		{/snippet}
	</Popup>
</div>
