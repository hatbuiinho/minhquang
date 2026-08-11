<script lang="ts">
	import Cropper from 'cropperjs';
	import 'cropperjs/dist/cropper.css';
	import { setScrollLock } from './scroll-lock';
	import { portal } from './portal';

	let {
		open,
		file,
		busy = false,
		onclose,
		oncrop
	}: {
		open: boolean;
		file: File | null;
		busy?: boolean;
		onclose: () => void;
		oncrop: (file: File) => void | Promise<void>;
	} = $props();
	const lockId = Symbol('avatar-crop');
	let objectUrl = $state('');
	let cropper: Cropper | null = null;
	let processing = $state(false);

	$effect(() => {
		setScrollLock(lockId, open);
		return () => setScrollLock(lockId, false);
	});

	$effect(() => {
		if (!open || !file) {
			objectUrl = '';
			cropper?.destroy();
			cropper = null;
			return;
		}
		const url = URL.createObjectURL(file);
		objectUrl = url;
		return () => URL.revokeObjectURL(url);
	});

	function mountCropper(node: HTMLImageElement) {
		const initialize = () => {
			cropper?.destroy();
			cropper = new Cropper(node, {
				aspectRatio: 1,
				viewMode: 1,
				dragMode: 'move',
				autoCropArea: 0.9,
				background: false,
				toggleDragModeOnDblclick: false
			});
		};
		if (node.complete) initialize();
		else node.addEventListener('load', initialize, { once: true });
		return { destroy: () => node.removeEventListener('load', initialize) };
	}

	async function crop() {
		if (!cropper || !file || processing || busy) return;
		processing = true;
		try {
			const canvas = cropper.getCroppedCanvas({
				width: 1080,
				height: 1080,
				imageSmoothingEnabled: true,
				imageSmoothingQuality: 'high'
			});
			const blob = await new Promise<Blob>((resolve, reject) =>
				canvas.toBlob(
					(value) => (value ? resolve(value) : reject(new Error('Không thể xử lý ảnh'))),
					'image/jpeg',
					0.92
				)
			);
			await oncrop(
				new File([blob], 'avatar.jpg', { type: 'image/jpeg', lastModified: Date.now() })
			);
		} finally {
			processing = false;
		}
	}
</script>

<svelte:window
	onkeydown={(event) => open && event.key === 'Escape' && !busy && !processing && onclose()}
/>

{#if open && file}
	<div
		use:portal
		class="fixed inset-0 z-[70] grid bg-black/60 sm:place-items-center sm:p-4"
		role="presentation"
	>
		<div
			class="flex h-dvh w-full flex-col overflow-hidden bg-[var(--color-surface)] sm:h-auto sm:max-h-[90dvh] sm:max-w-3xl sm:rounded-md"
			role="dialog"
			aria-modal="true"
			aria-label="Cắt ảnh đại diện"
		>
			<header
				class="flex items-center justify-between border-b border-[var(--color-border)] px-4 py-3"
			>
				<div>
					<h2 class="font-semibold">Cắt ảnh đại diện</h2>
					<p class="text-sm text-[var(--color-text-secondary)]">
						Ảnh sẽ được lưu theo tỉ lệ vuông 1:1
					</p>
				</div>
				<button
					type="button"
					disabled={busy || processing}
					class="grid h-10 w-10 place-items-center rounded-md hover:bg-[var(--color-surface-muted)]"
					aria-label="Đóng"
					onclick={onclose}
					><span class="icon-[lucide--x] h-5 w-5" aria-hidden="true"></span></button
				>
			</header>
			<div class="min-h-0 flex-1 bg-neutral-950 p-3 sm:p-5">
				<div
					class="flex h-full min-h-72 items-center justify-center overflow-hidden bg-neutral-900"
				>
					{#if objectUrl}<img
							src={objectUrl}
							alt="Ảnh cần cắt"
							class="block max-h-full max-w-full object-contain"
							use:mountCropper
						/>{/if}
				</div>
			</div>
			<footer
				class="flex justify-end gap-3 border-t border-[var(--color-border)] px-3 pt-3 pb-[max(env(safe-area-inset-bottom),0.75rem)] sm:p-3"
			>
				<button
					type="button"
					disabled={busy || processing}
					class="h-11 rounded-md border border-[var(--color-border-strong)] px-5 font-semibold"
					onclick={onclose}>Huỷ</button
				>
				<button
					type="button"
					disabled={busy || processing}
					class="flex h-11 items-center gap-2 rounded-md bg-[var(--color-primary)] px-5 font-semibold text-white disabled:opacity-50"
					onclick={() => void crop()}
					>{#if busy || processing}<span
							class="icon-[lucide--loader-circle] h-4 w-4 animate-spin"
							aria-hidden="true"
						></span>{/if}Dùng ảnh này</button
				>
			</footer>
		</div>
	</div>
{/if}
