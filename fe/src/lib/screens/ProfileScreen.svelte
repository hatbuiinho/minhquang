<script module lang="ts">
	function formatDate(value: string) {
		return new Intl.DateTimeFormat('vi-VN', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric'
		}).format(new Date(value));
	}
</script>

<script lang="ts">
	import { tick } from 'svelte';
	import { authStore } from '$lib/auth/auth-store.svelte';
	import { userStore } from '$lib/users/user-store.svelte';
	import AvatarUploader from '$lib/ui/AvatarUploader.svelte';
	import { uploadAvatar } from '$lib/uploads/api';
	import { toastStore } from '$lib/ui/toast-store.svelte';

	let user = $derived(authStore.user);
	let editing = $state(false);
	let displayName = $state('');
	let username = $state('');
	let displayNameInput = $state<HTMLInputElement>();
	let avatarUploading = $state(false);

	async function startEditing() {
		displayName = user?.display_name ?? '';
		username = user?.username ?? '';
		editing = true;
		await tick();
		displayNameInput?.focus();
		displayNameInput?.setSelectionRange(displayName.length, displayName.length);
	}

	function cancelEditing() {
		editing = false;
	}

	async function saveProfile(event: SubmitEvent) {
		event.preventDefault();
		const item = await authStore.updateProfile(username, displayName);
		if (!item) return;
		userStore.sync(item);
		editing = false;
	}

	async function selectAvatar(file: File) {
		avatarUploading = true;
		try {
			const avatarUrl = await uploadAvatar(file);
			const item = await authStore.updateAvatar(avatarUrl);
			if (item) userStore.sync(item);
		} catch (error) {
			toastStore.error(error instanceof Error ? error.message : 'Không thể tải ảnh lên');
		} finally {
			avatarUploading = false;
		}
	}
</script>

<section class="h-full overflow-y-auto px-4 py-5 md:px-6 md:py-6 lg:px-8">
	<div class="mx-auto max-w-3xl">
		<div class="border-b border-[var(--color-border)] pb-5">
			<AvatarUploader
				avatarUrl={user?.avatar_url}
				displayName={user?.display_name ?? ''}
				uploading={avatarUploading}
				onselect={selectAvatar}
			/>
			<div class="mt-5 flex items-center gap-4">
				<div class="min-w-0 flex-1">
					<h1 class="truncate text-lg font-semibold">{user?.display_name}</h1>
					<p class="mt-1 truncate text-sm text-[var(--color-text-secondary)]">@{user?.username}</p>
				</div>
				{#if !editing}
					<button
						type="button"
						class="grid h-10 w-10 shrink-0 place-items-center rounded-md border border-[var(--color-border-strong)] text-[var(--color-primary-dark)] hover:bg-[var(--color-surface-muted)]"
						aria-label="Chỉnh sửa hồ sơ"
						title="Chỉnh sửa hồ sơ"
						onclick={() => void startEditing()}
					>
						<span class="icon-[lucide--square-pen] h-5 w-5" aria-hidden="true"></span>
					</button>
				{/if}
			</div>
		</div>

		{#if editing}
			<form
				class="grid gap-4 border-b border-[var(--color-border)] py-5 sm:grid-cols-2"
				onsubmit={saveProfile}
			>
				<label class="block">
					<span class="mb-1.5 block text-sm font-medium">Tên hiển thị</span>
					<input
						bind:this={displayNameInput}
						bind:value={displayName}
						required
						autocomplete="name"
						class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
					/>
				</label>
				<label class="block">
					<span class="mb-1.5 block text-sm font-medium">Tên đăng nhập</span>
					<input
						bind:value={username}
						required
						autocomplete="username"
						class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
					/>
				</label>
				<div class="grid grid-cols-2 gap-3 sm:col-span-2 sm:flex sm:justify-end">
					<button
						type="button"
						disabled={authStore.isUpdatingProfile}
						class="h-11 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-5 text-sm font-semibold disabled:opacity-50"
						onclick={cancelEditing}>Huỷ</button
					>
					<button
						type="submit"
						disabled={authStore.isUpdatingProfile}
						class="flex h-11 items-center justify-center gap-2 rounded-md bg-[var(--color-primary)] px-5 text-sm font-semibold text-white disabled:opacity-50"
					>
						{#if authStore.isUpdatingProfile}
							<span class="icon-[lucide--loader-circle] h-4 w-4 animate-spin" aria-hidden="true"
							></span>
						{/if}
						Lưu
					</button>
				</div>
			</form>
		{/if}

		<dl class="divide-y divide-[var(--color-border)]">
			{#if !editing}
				<div class="grid gap-1 py-4 sm:grid-cols-[11rem_1fr] sm:items-center">
					<dt class="text-sm text-[var(--color-text-secondary)]">Tên hiển thị</dt>
					<dd class="text-sm font-medium">{user?.display_name}</dd>
				</div>
				<div class="grid gap-1 py-4 sm:grid-cols-[11rem_1fr] sm:items-center">
					<dt class="text-sm text-[var(--color-text-secondary)]">Tên đăng nhập</dt>
					<dd class="text-sm font-medium">@{user?.username}</dd>
				</div>
			{/if}
			<div class="grid gap-1 py-4 sm:grid-cols-[11rem_1fr] sm:items-center">
				<dt class="text-sm text-[var(--color-text-secondary)]">Vai trò</dt>
				<dd class="text-sm font-medium">
					{user?.role === 'admin' ? 'Quản trị viên' : 'Giám sát viên'}
				</dd>
			</div>
			<div class="grid gap-1 py-4 sm:grid-cols-[11rem_1fr] sm:items-center">
				<dt class="text-sm text-[var(--color-text-secondary)]">Trạng thái</dt>
				<dd class="text-sm font-medium text-[var(--color-primary-dark)]">
					{user?.active ? 'Đang hoạt động' : 'Ngừng hoạt động'}
				</dd>
			</div>
			{#if user?.created_at}
				<div class="grid gap-1 py-4 sm:grid-cols-[11rem_1fr] sm:items-center">
					<dt class="text-sm text-[var(--color-text-secondary)]">Ngày tạo tài khoản</dt>
					<dd class="text-sm font-medium">{formatDate(user.created_at)}</dd>
				</div>
			{/if}
		</dl>
	</div>
</section>
