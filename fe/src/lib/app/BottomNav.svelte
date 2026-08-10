<script lang="ts">
	import { bottomNavItems, mainRouteFor, type AppRoute } from '$lib/navigation/routes';
	import { router } from '$lib/navigation/router.svelte';

	let { route }: { route: AppRoute } = $props();
	let active = $derived(mainRouteFor(route));
</script>

<nav
	class="z-20 border-t border-[var(--color-border)] bg-[rgb(255_255_255_/_0.94)] px-2 pt-2 pb-[max(env(safe-area-inset-bottom),0.5rem)] backdrop-blur"
>
	<div class="grid grid-cols-4 gap-1">
		{#each bottomNavItems as item (item.name)}
			<button
				type="button"
				class={[
					'flex h-14 flex-col items-center justify-center gap-1 rounded-[12px] text-xs font-medium transition',
					active === item.name
						? 'bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]'
						: 'text-[var(--color-text-muted)] hover:bg-[var(--color-surface-muted)] hover:text-[var(--color-text-secondary)]'
				]}
				aria-current={active === item.name ? 'page' : undefined}
				onclick={() => router.openMain(item.name)}
			>
				<span class={['h-5 w-5', item.icon]} aria-hidden="true"></span>
				<span>{item.label}</span>
			</button>
		{/each}
	</div>
</nav>
