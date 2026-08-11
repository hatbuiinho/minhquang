<script lang="ts">
	import { setScrollLock } from './scroll-lock';
	import { portal } from './portal';

	let {
		src,
		alt = '',
		open = $bindable(false)
	}: { src: string; alt?: string; open?: boolean } = $props();
	const lockId = Symbol('lightbox');
	let scale = $state(1);
	let rotation = $state(0);
	let x = $state(0);
	let y = $state(0);
	let dragging = $state(false);
	let startX = 0;
	let startY = 0;
	let originX = 0;
	let originY = 0;

	$effect(() => {
		setScrollLock(lockId, open);
		if (!open) reset();
		return () => setScrollLock(lockId, false);
	});

	function reset() {
		scale = 1;
		rotation = 0;
		x = 0;
		y = 0;
	}

	function close() {
		open = false;
	}

	function zoom(delta: number) {
		scale = Math.min(4, Math.max(1, scale + delta));
		if (scale === 1) {
			x = 0;
			y = 0;
		}
	}

	function pointerDown(event: PointerEvent) {
		if (scale === 1) return;
		dragging = true;
		startX = event.clientX;
		startY = event.clientY;
		originX = x;
		originY = y;
		(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
	}

	function pointerMove(event: PointerEvent) {
		if (!dragging) return;
		x = originX + event.clientX - startX;
		y = originY + event.clientY - startY;
	}
</script>

<svelte:window
	onkeydown={(event) => {
		if (!open) return;
		if (event.key === 'Escape') close();
		if (event.key === '+' || event.key === '=') zoom(0.5);
		if (event.key === '-') zoom(-0.5);
	}}
/>

{#if open && src}
	<div
		use:portal
		class="fixed inset-0 z-[80] bg-black/95 text-white"
		role="dialog"
		aria-modal="true"
		aria-label="Xem ảnh"
		tabindex="-1"
		onkeydown={(event) => event.key === 'Escape' && close()}
	>
		<div class="absolute top-[max(env(safe-area-inset-top),0.75rem)] right-3 z-10 flex gap-2">
			<button
				type="button"
				class="grid h-10 w-10 place-items-center rounded-full bg-white/10 hover:bg-white/20"
				aria-label="Xoay ảnh"
				onclick={() => (rotation = (rotation - 90) % 360)}
			>
				<span class="icon-[lucide--rotate-ccw] h-5 w-5" aria-hidden="true"></span>
			</button>
			<button
				type="button"
				disabled={scale === 1}
				class="grid h-10 w-10 place-items-center rounded-full bg-white/10 hover:bg-white/20 disabled:opacity-40"
				aria-label="Thu nhỏ"
				onclick={() => zoom(-0.5)}
			>
				<span class="icon-[lucide--zoom-out] h-5 w-5" aria-hidden="true"></span>
			</button>
			<button
				type="button"
				disabled={scale === 4}
				class="grid h-10 w-10 place-items-center rounded-full bg-white/10 hover:bg-white/20 disabled:opacity-40"
				aria-label="Phóng to"
				onclick={() => zoom(0.5)}
			>
				<span class="icon-[lucide--zoom-in] h-5 w-5" aria-hidden="true"></span>
			</button>
			<button
				type="button"
				class="grid h-10 w-10 place-items-center rounded-full bg-white/10 hover:bg-white/20"
				aria-label="Đóng"
				onclick={close}
			>
				<span class="icon-[lucide--x] h-5 w-5" aria-hidden="true"></span>
			</button>
		</div>
		<div
			class="flex h-full w-full touch-none items-center justify-center overflow-hidden pt-16 pb-4"
			role="presentation"
			onclick={(event) => event.target === event.currentTarget && close()}
			ondblclick={() => (scale > 1 ? reset() : zoom(1))}
			onwheel={(event) => {
				event.preventDefault();
				zoom(event.deltaY < 0 ? 0.25 : -0.25);
			}}
			onpointerdown={pointerDown}
			onpointermove={pointerMove}
			onpointerup={() => (dragging = false)}
			onpointercancel={() => (dragging = false)}
		>
			<img
				{src}
				{alt}
				draggable="false"
				class="max-h-full max-w-full object-contain transition-transform duration-150 select-none"
				class:cursor-grab={scale > 1 && !dragging}
				class:cursor-grabbing={dragging}
				style={`transform: translate3d(${x}px, ${y}px, 0) scale(${scale}) rotate(${rotation}deg);`}
			/>
		</div>
	</div>
{/if}
