import { apiRequest } from '$lib/api/client';

type PresignedUpload = {
	bucket: string;
	object_key: string;
	upload_url: string;
	public_url: string;
	expires_at: string;
};

export async function uploadAvatar(file: File): Promise<string> {
	const presigned = await apiRequest<PresignedUpload>('/api/uploads/presign', {
		method: 'POST',
		body: JSON.stringify({ file_name: file.name, content_type: file.type, kind: 'avatar' })
	});
	const response = await fetch(presigned.upload_url, {
		method: 'PUT',
		headers: { 'Content-Type': file.type },
		body: file
	});
	if (!response.ok) throw new Error(`Không thể tải ảnh lên (${response.status})`);
	return presigned.public_url;
}
