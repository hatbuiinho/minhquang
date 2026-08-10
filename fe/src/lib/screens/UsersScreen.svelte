<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/api/client';
	import type { AdminUser } from '$lib/auth/auth-store.svelte';
	import { authStore } from '$lib/auth/auth-store.svelte';
	import { toastStore } from '$lib/ui/toast-store.svelte';

	let users = $state<AdminUser[]>([]);
	let showForm = $state(false);
	let saving = $state(false);
	let displayName = $state('');
	let username = $state('');
	let password = $state('');

	onMount(load);

	async function load() {
		try {
			users = (await apiRequest<{ users: AdminUser[] }>('/api/users')).users;
		} catch (error) {
			toastStore.error(error instanceof Error ? error.message : 'Không thể tải tài khoản');
		}
	}

	async function create(event: SubmitEvent) {
		event.preventDefault();
		saving = true;
		try {
			await apiRequest<AdminUser>('/api/users', {
				method: 'POST',
				body: JSON.stringify({ username, display_name: displayName, password })
			});
			toastStore.success('Đã tạo tài khoản quản trị');
			displayName = '';
			username = '';
			password = '';
			showForm = false;
			await load();
		} catch (error) {
			toastStore.error(error instanceof Error ? error.message : 'Không thể tạo tài khoản');
		} finally {
			saving = false;
		}
	}
</script>

<section class="h-full overflow-y-auto px-4 py-4 md:px-6 md:py-6 lg:px-8">
	<div class="mx-auto max-w-[1200px]">
		<div class="mb-4 flex items-center justify-between gap-3">
			<div>
				<p class="text-sm font-medium">Tài khoản quản trị</p>
				<p class="mt-0.5 text-xs text-[var(--color-text-secondary)]">{users.length} tài khoản</p>
			</div>
			<button
				type="button"
				class="grid h-10 w-10 place-items-center rounded-md bg-[var(--color-primary)] text-white"
				aria-label="Thêm tài khoản"
				onclick={() => (showForm = !showForm)}
			>
				<span
					class={showForm ? 'icon-[lucide--x] h-5 w-5' : 'icon-[lucide--user-plus] h-5 w-5'}
					aria-hidden="true"
				></span>
			</button>
		</div>

		{#if showForm}
			<form
				class="mb-5 grid gap-3 rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] p-4 md:grid-cols-2 lg:grid-cols-4"
				onsubmit={create}
			>
				<input
					bind:value={displayName}
					required
					placeholder="Tên hiển thị"
					aria-label="Tên hiển thị"
					class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
				/>
				<input
					bind:value={username}
					required
					placeholder="Tên đăng nhập"
					aria-label="Tên đăng nhập"
					autocomplete="off"
					class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
				/>
				<input
					bind:value={password}
					required
					minlength="8"
					type="password"
					placeholder="Mật khẩu (ít nhất 8 ký tự)"
					aria-label="Mật khẩu"
					autocomplete="new-password"
					class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
				/>
				<button
					type="submit"
					disabled={saving}
					class="h-11 w-full rounded-md bg-[var(--color-primary)] text-sm font-semibold text-white disabled:opacity-60"
					>{saving ? 'Đang tạo...' : 'Tạo tài khoản'}</button
				>
			</form>
		{/if}

		<ul class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
			{#each users as user (user.id)}
				<li
					class="flex items-center gap-3 rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] p-3"
				>
					<span
						class="grid h-10 w-10 place-items-center rounded-full bg-[var(--color-primary-soft)] text-sm font-semibold text-[var(--color-primary-dark)]"
						>{user.display_name.slice(0, 1).toUpperCase()}</span
					>
					<div class="min-w-0 flex-1">
						<p class="truncate text-sm font-semibold">{user.display_name}</p>
						<p class="truncate text-xs text-[var(--color-text-secondary)]">@{user.username}</p>
					</div>
					{#if user.id === authStore.user?.id}<span class="text-xs text-[var(--color-primary)]"
							>Bạn</span
						>{/if}
				</li>
			{/each}
		</ul>
	</div>
</section>
