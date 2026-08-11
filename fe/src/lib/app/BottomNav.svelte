<script lang="ts">
	import { authStore } from '$lib/auth/auth-store.svelte';
	import { bottomNavItems, mainRouteFor, type AppRoute } from '$lib/navigation/routes';
	import { router } from '$lib/navigation/router.svelte';
	let { route }: { route: AppRoute } = $props();
	let active = $derived(mainRouteFor(route));
	let items = $derived(
		bottomNavItems.filter((item) => item.name !== 'users' || authStore.can('user.read'))
	);
</script>

<nav
	class="z-20 border-t border-[var(--color-border)] bg-[rgb(255_255_255_/_0.96)] px-2 pt-2 pb-[max(env(safe-area-inset-bottom),0.5rem)] backdrop-blur md:hidden"
>
	<div class={['grid gap-1', items.length === 2 ? 'grid-cols-2' : 'grid-cols-3']}>
		{#each items as item (item.name)}
			<button
				type="button"
				class={[
					'flex h-14 flex-col items-center justify-center gap-1 rounded-md text-xs font-medium',
					active === item.name
						? 'bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]'
						: 'text-[var(--color-text-muted)]'
				]}
				aria-current={active === item.name ? 'page' : undefined}
				onclick={() => router.openMain(item.name)}
			>
				<span class={['h-5 w-5', item.icon]} aria-hidden="true"></span><span>{item.label}</span>
			</button>
		{/each}
	</div>
</nav>
