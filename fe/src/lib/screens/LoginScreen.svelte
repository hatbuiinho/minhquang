<script lang="ts">
	import Logo from '$lib/ui/Logo.svelte';
	import { authStore } from '$lib/auth/auth-store.svelte';

	let username = $state('');
	let password = $state('');

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		await authStore.login(username, password);
	}
</script>

<section class="flex min-h-screen items-center justify-center px-5 py-10">
	<div class="w-full max-w-sm">
		<div class="mb-8 text-center">
			<Logo />
			<h1 class="mt-5 text-2xl font-semibold">Quản lý nhân sự công quả</h1>
			<p class="mt-2 text-sm text-[var(--color-text-secondary)]">Thiền Viện Minh Quang</p>
		</div>

		<form
			class="space-y-4 rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)] p-5 shadow-[var(--shadow-soft)]"
			onsubmit={submit}
		>
			<label class="block">
				<span class="mb-1.5 block text-sm font-medium">Tên đăng nhập</span>
				<input
					bind:value={username}
					autocomplete="username"
					required
					class="h-12 w-full rounded-md border-[var(--color-border-strong)]"
				/>
			</label>
			<label class="block">
				<span class="mb-1.5 block text-sm font-medium">Mật khẩu</span>
				<input
					bind:value={password}
					type="password"
					autocomplete="current-password"
					required
					class="h-12 w-full rounded-md border-[var(--color-border-strong)]"
				/>
			</label>
			{#if authStore.error}
				<p class="text-sm text-[var(--color-danger)]">{authStore.error}</p>
			{/if}
			<button
				type="submit"
				disabled={authStore.isSubmitting}
				class="flex h-12 w-full items-center justify-center gap-2 rounded-md bg-[var(--color-primary)] text-sm font-semibold text-white disabled:opacity-60"
			>
				<span class="icon-[lucide--log-in] h-5 w-5" aria-hidden="true"></span>
				{authStore.isSubmitting ? 'Đang đăng nhập...' : 'Đăng nhập'}
			</button>
		</form>
	</div>
</section>
