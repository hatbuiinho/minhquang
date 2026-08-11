<script lang="ts">
	import { tick } from 'svelte';
	import { toastStore } from './toast-store.svelte';
	import AvatarCropDialog from './AvatarCropDialog.svelte';
	import Lightbox from './Lightbox.svelte';
	import { portal } from './portal';

	let {
		avatarUrl = '',
		displayName,
		uploading = false,
		editable = true,
		onselect
	}: {
		avatarUrl?: string;
		displayName: string;
		uploading?: boolean;
		editable?: boolean;
		onselect: (file: File) => void | Promise<void>;
	} = $props();
	let libraryInput = $state<HTMLInputElement>();
	let cameraInput = $state<HTMLInputElement>();
	let menuOpen = $state(false);
	let cropOpen = $state(false);
	let cropFile = $state<File | null>(null);
	let lightboxOpen = $state(false);
	let cameraButton = $state<HTMLButtonElement>();
	let menuElement = $state<HTMLDivElement>();
	let menuLeft = $state(0);
	let menuTop = $state(0);
	let menuPositioned = $state(false);
	let initial = $derived(displayName.trim().slice(0, 1).toUpperCase() || '?');

	$effect(() => {
		if (!menuOpen) return;
		void tick().then(positionMenu);
		window.addEventListener('resize', positionMenu);
		document.addEventListener('scroll', positionMenu, true);
		return () => {
			window.removeEventListener('resize', positionMenu);
			document.removeEventListener('scroll', positionMenu, true);
		};
	});

	async function toggleMenu() {
		menuOpen = !menuOpen;
		menuPositioned = false;
		if (!menuOpen) return;
		await tick();
		positionMenu();
	}

	function positionMenu() {
		if (!menuOpen || !cameraButton || !menuElement) return;
		const margin = 8;
		const gap = 8;
		const anchor = cameraButton.getBoundingClientRect();
		const menuWidth = menuElement.offsetWidth;
		const menuHeight = menuElement.offsetHeight;

		let left = anchor.right - menuWidth;
		if (left < margin) left = anchor.left;
		left = Math.min(Math.max(margin, left), window.innerWidth - menuWidth - margin);

		let top = anchor.bottom + gap;
		if (top + menuHeight > window.innerHeight - margin) top = anchor.top - menuHeight - gap;
		top = Math.min(Math.max(margin, top), window.innerHeight - menuHeight - margin);

		menuLeft = left;
		menuTop = top;
		menuPositioned = true;
	}

	function choose(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (!file) return;
		if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) {
			toastStore.error('Chỉ hỗ trợ ảnh JPEG, PNG hoặc WebP');
			return;
		}
		if (file.size > 15 * 1024 * 1024) {
			toastStore.error('Ảnh gốc không được lớn hơn 15 MB');
			return;
		}
		cropFile = file;
		cropOpen = true;
	}

	async function cropped(file: File) {
		await onselect(file);
		cropOpen = false;
		cropFile = null;
	}
</script>

<div class="flex items-start gap-4">
	<div class="relative shrink-0">
		<button
			type="button"
			class="relative grid h-24 w-24 place-items-center overflow-hidden rounded-full border border-[var(--color-border-strong)] bg-[var(--color-primary-soft)] text-2xl font-semibold text-[var(--color-primary-dark)]"
			aria-label={avatarUrl ? 'Xem ảnh đại diện' : 'Ảnh đại diện'}
			onclick={() => avatarUrl && (lightboxOpen = true)}
		>
			{#if avatarUrl}<img
					src={avatarUrl}
					alt={displayName}
					class="h-full w-full object-cover"
				/>{:else}{initial}{/if}
			{#if uploading}<span class="absolute inset-0 grid place-items-center bg-black/45 text-white"
					><span class="icon-[lucide--loader-circle] h-6 w-6 animate-spin" aria-hidden="true"
					></span></span
				>{/if}
		</button>
		{#if editable}<div class="absolute right-0 bottom-0">
				<button
					bind:this={cameraButton}
					type="button"
					disabled={uploading}
					class="grid h-10 w-10 place-items-center rounded-full border-2 border-white bg-[var(--color-primary)] text-white shadow"
					aria-label="Đổi ảnh đại diện"
					onclick={() => void toggleMenu()}
					><span class="icon-[lucide--camera] h-4.5 w-4.5" aria-hidden="true"></span></button
				>
				{#if menuOpen}
					<button
						use:portal
						type="button"
						class="fixed inset-0 z-[59] cursor-default"
						aria-label="Đóng menu ảnh"
						onclick={() => (menuOpen = false)}
					></button>
					<div
						use:portal
						bind:this={menuElement}
						class="fixed z-[60] w-[min(13rem,calc(100vw-1rem))] overflow-hidden rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] py-1 shadow-[var(--shadow-popover)]"
						style={`left: ${menuLeft}px; top: ${menuTop}px; visibility: ${menuPositioned ? 'visible' : 'hidden'};`}
					>
						<button
							type="button"
							class="flex h-11 w-full items-center gap-3 px-3 text-sm hover:bg-[var(--color-surface-muted)]"
							onclick={() => {
								menuOpen = false;
								cameraInput?.click();
							}}
							><span class="icon-[lucide--camera] h-4.5 w-4.5" aria-hidden="true"></span>Chụp ảnh</button
						>
						<button
							type="button"
							class="flex h-11 w-full items-center gap-3 px-3 text-sm hover:bg-[var(--color-surface-muted)]"
							onclick={() => {
								menuOpen = false;
								libraryInput?.click();
							}}
							><span class="icon-[lucide--image] h-4.5 w-4.5" aria-hidden="true"></span>Chọn từ thư
							viện</button
						>
					</div>
				{/if}
			</div>{/if}
	</div>
	<div class="min-w-0 pt-1">
		<p class="text-sm font-semibold">Ảnh đại diện</p>
		{#if editable}<p class="mt-1 text-sm leading-5 text-[var(--color-text-secondary)]">
				Chọn ảnh rõ khuôn mặt. Ảnh được cắt vuông trước khi tải lên.
			</p>{/if}
		{#if avatarUrl}<button
				type="button"
				class="mt-2 text-sm font-semibold text-[var(--color-primary)]"
				onclick={() => (lightboxOpen = true)}>Xem ảnh</button
			>{/if}
	</div>
	<input
		bind:this={libraryInput}
		type="file"
		accept="image/jpeg,image/png,image/webp"
		class="hidden"
		onchange={choose}
	/>
	<input
		bind:this={cameraInput}
		type="file"
		accept="image/jpeg,image/png,image/webp"
		capture="environment"
		class="hidden"
		onchange={choose}
	/>
</div>

<Lightbox src={avatarUrl} alt={displayName} bind:open={lightboxOpen} />
<AvatarCropDialog
	open={cropOpen}
	file={cropFile}
	busy={uploading}
	onclose={() => {
		if (!uploading) {
			cropOpen = false;
			cropFile = null;
		}
	}}
	oncrop={cropped}
/>
