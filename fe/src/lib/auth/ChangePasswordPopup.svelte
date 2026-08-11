<script lang="ts">
	import { authStore } from './auth-store.svelte';
	import { changePasswordPopupStore } from './change-password-popup-store.svelte';
	import PasswordInput from '$lib/ui/PasswordInput.svelte';
	import Popup from '$lib/ui/Popup.svelte';

	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordError = $state('');

	$effect(() => {
		if (!changePasswordPopupStore.open) return;
		currentPassword = '';
		newPassword = '';
		confirmPassword = '';
		passwordError = '';
	});

	async function changePassword(event: SubmitEvent) {
		event.preventDefault();
		passwordError = '';
		if (newPassword !== confirmPassword) {
			passwordError = 'Mật khẩu xác nhận không khớp';
			return;
		}
		if (!(await authStore.changePassword(currentPassword, newPassword))) return;
		changePasswordPopupStore.close();
	}
</script>

<Popup
	open={changePasswordPopupStore.open}
	title="Đổi mật khẩu"
	onClose={() => {
		if (!authStore.isChangingPassword) changePasswordPopupStore.close();
	}}
>
	<form id="change-password-form" class="space-y-4" onsubmit={changePassword}>
		<div>
			<span class="mb-1.5 block text-sm font-medium">Mật khẩu hiện tại</span>
			<PasswordInput
				bind:value={currentPassword}
				required
				autocomplete="current-password"
				ariaLabel="Mật khẩu hiện tại"
			/>
		</div>
		<div>
			<span class="mb-1.5 block text-sm font-medium">Mật khẩu mới</span>
			<PasswordInput
				bind:value={newPassword}
				required
				minlength={8}
				autocomplete="new-password"
				ariaLabel="Mật khẩu mới"
			/>
		</div>
		<div>
			<span class="mb-1.5 block text-sm font-medium">Xác nhận mật khẩu mới</span>
			<PasswordInput
				bind:value={confirmPassword}
				required
				minlength={8}
				autocomplete="new-password"
				ariaLabel="Xác nhận mật khẩu mới"
			/>
		</div>
		{#if passwordError}
			<p class="text-sm text-[var(--color-danger)]">{passwordError}</p>
		{/if}
	</form>

	{#snippet footer()}
		<div class="grid grid-cols-2 gap-3">
			<button
				type="button"
				disabled={authStore.isChangingPassword}
				class="h-11 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] text-sm font-semibold disabled:opacity-50"
				onclick={() => changePasswordPopupStore.close()}>Huỷ</button
			>
			<button
				type="submit"
				form="change-password-form"
				disabled={authStore.isChangingPassword}
				class="flex h-11 items-center justify-center gap-2 rounded-md bg-[var(--color-primary)] text-sm font-semibold text-white disabled:opacity-50"
			>
				{#if authStore.isChangingPassword}
					<span class="icon-[lucide--loader-circle] h-4 w-4 animate-spin" aria-hidden="true"></span>
				{/if}
				Đổi mật khẩu
			</button>
		</div>
	{/snippet}
</Popup>
