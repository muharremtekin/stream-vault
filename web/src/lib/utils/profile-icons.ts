export const PROFILE_ICONS = [
  { id: 'smile', color: '#e50914' },
  { id: 'cat', color: '#3b82f6' },
  { id: 'dog', color: '#22c55e' },
  { id: 'bird', color: '#f59e0b' },
  { id: 'fish', color: '#8b5cf6' },
  { id: 'rabbit', color: '#ec4899' },
  { id: 'star', color: '#06b6d4' },
  { id: 'heart', color: '#ef4444' },
  { id: 'ghost', color: '#a3a3a3' },
  { id: 'rocket', color: '#f97316' },
] as const;

export type ProfileIconId = (typeof PROFILE_ICONS)[number]['id'];

export function getProfileIcon(id: string) {
  return PROFILE_ICONS.find((icon) => icon.id === id) ?? PROFILE_ICONS[0];
}
