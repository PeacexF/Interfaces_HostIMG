import { redirect, fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { bffFetch } from '$lib/server/bff';

export const load: PageServerLoad = async ({ parent, request }) => {
	const { accountId } = await parent();
	if (!accountId) redirect(303, '/login');

	const cookie = request.headers.get('cookie') ?? '';
	const res = await bffFetch('/api/files?limit=50', cookie);
	if (!res.ok) return { files: [] };
	const data = (await res.json()) as {
		files: Array<{ id: string; name: string; size: number; mime_type: string; created_at: string }>;
	};
	return { files: data.files };
};

export const actions: Actions = {
	upload: async ({ request }) => {
		const cookie = request.headers.get('cookie') ?? '';
		const form = await request.formData();
		const res = await bffFetch('/api/files', cookie, { method: 'POST', body: form });
		if (!res.ok) {
			const body = await res.json().catch(() => ({ error: 'upload failed' }));
			return fail(res.status, { error: body.error as string });
		}
		return { success: true };
	},
	delete: async ({ request }) => {
		const cookie = request.headers.get('cookie') ?? '';
		const form = await request.formData();
		const id = form.get('id');
		const res = await bffFetch(`/api/files/${id}`, cookie, { method: 'DELETE' });
		if (!res.ok) return fail(res.status, { error: 'delete failed' });
		return { success: true };
	}
};