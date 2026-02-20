import { z } from 'zod';

const channelPreferenceSchema = z.object({
  enabled: z.boolean(),
  categories: z.array(z.string()).optional(),
});

export const notificationPreferencesSchema = z.object({
  email: channelPreferenceSchema,
  push: channelPreferenceSchema,
  inApp: channelPreferenceSchema,
});

export type NotificationPreferencesFormData = z.infer<typeof notificationPreferencesSchema>;
