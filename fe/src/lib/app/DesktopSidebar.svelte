<script lang="ts">
	import { authStore } from '$lib/auth/auth-store.svelte';
	import { bottomNavItems, mainRouteFor, type AppRoute } from '$lib/navigation/routes';
	import { router } from '$lib/navigation/router.svelte';
	import Logo from '$lib/ui/Logo.svelte';

	let { route }: { route: AppRoute } = $props();
	let active = $derived(mainRouteFor(route));
</script>

<aside
	class="hidden h-dvh w-60 shrink-0 flex-col border-r border-[var(--color-border)] bg-[var(--color-surface)] md:flex"
>
	<div class="flex h-[73px] items-center border-b border-[var(--color-border)] px-5">
		<Logo />
	</div>

	<nav class="flex-1 space-y-1 px-3 py-5" aria-label="Điều hướng chính">
		{#each bottomNavItems as item (item.name)}
			<button
				type="button"
				class={[
					'flex h-11 w-full items-center gap-3 rounded-md px-3 text-sm font-semibold',
					active === item.name
						? 'bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]'
						: 'text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-muted)]'
				]}
				aria-current={active === item.name ? 'page' : undefined}
				onclick={() => router.openMain(item.name)}
			>
				<span class={['h-5 w-5 shrink-0', item.icon]} aria-hidden="true"></span>
				<span>{item.label}</span>
			</button>
		{/each}
	</nav>

	<div class="border-t border-[var(--color-border)] p-3">
		<div class="mb-2 flex min-w-0 items-center gap-3 px-2 py-2">
			<span
				class="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-[var(--color-primary-soft)] text-sm font-semibold text-[var(--color-primary-dark)]"
			>
				{authStore.user?.display_name.slice(0, 1).toUpperCase()}
			</span>
			<div class="min-w-0">
				<p class="truncate text-sm font-semibold">{authStore.user?.display_name}</p>
				<p class="truncate text-xs text-[var(--color-text-secondary)]">
					@{authStore.user?.username}
				</p>
			</div>
		</div>
		<button
			type="button"
			class="flex h-10 w-full items-center gap-3 rounded-md px-3 text-sm font-medium text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-muted)]"
			onclick={() => authStore.logout()}
		>
			<span class="icon-[lucide--log-out] h-4 w-4" aria-hidden="true"></span>
			Đăng xuất
		</button>
	</div>
</aside>
