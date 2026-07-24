import type { LayoutServerLoad } from './$types';
import { bffFetch } from '$lib/server/bff';

export const load: LayoutServerLoad = async ({ request }) => {
	const cookie = request.headers.get('cookie') ?? '';
	const res = await bffFetch('/api/me', cookie);
	if (!res.ok) return { accountId: null };
	const data = (await res.json()) as { account_id: number };
	return { accountId: data.account_id };
};