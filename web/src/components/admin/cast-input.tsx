'use client';

import { useRef } from 'react';
import { useTranslations } from 'next-intl';
import { Plus, Trash2 } from 'lucide-react';

import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

interface CastMemberField {
  id?: string;
  name: string;
  role: string;
  photoUrl?: string | null;
}

interface CastMemberWithId extends CastMemberField {
  id: string;
}

interface CastInputProps {
  value: CastMemberField[];
  onChange: (cast: CastMemberField[]) => void;
  error?: string;
}

export function CastInput({ value, onChange, error }: CastInputProps) {
  const idMapRef = useRef(new Map<number, string>());
  const t = useTranslations('admin.contentForm');

  const members = value.map((m, i) => {
    if (m.id) return m as CastMemberWithId;
    const existing = idMapRef.current.get(i);
    if (existing) return { ...m, id: existing };
    const newId = crypto.randomUUID();
    idMapRef.current.set(i, newId);
    return { ...m, id: newId };
  });

  const addMember = () => {
    onChange([...members, { id: crypto.randomUUID(), name: '', role: '' }]);
  };

  const removeMember = (index: number) => {
    onChange(members.filter((_, i) => i !== index));
  };

  const updateMember = (index: number, field: 'name' | 'role', val: string) => {
    const updated = members.map((m, i) =>
      i === index ? { ...m, [field]: val } : m
    );
    onChange(updated);
  };

  return (
    <div className="w-full">
      <label className="mb-1.5 block text-sm font-medium text-foreground">
        {t('cast')}
      </label>
      <div className="space-y-2">
        {members.map((member, i) => (
          <div key={member.id} className="flex items-start gap-2">
            <Input
              placeholder={t('castName')}
              value={member.name}
              onChange={(e) => updateMember(i, 'name', e.target.value)}
              size="sm"
              className="flex-1"
            />
            <Input
              placeholder={t('castRole')}
              value={member.role}
              onChange={(e) => updateMember(i, 'role', e.target.value)}
              size="sm"
              className="flex-1"
            />
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => removeMember(i)}
              aria-label={t('removeCast')}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        ))}
      </div>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={addMember}
        className="mt-2"
      >
        <Plus className="mr-1 h-4 w-4" />
        {t('addCast')}
      </Button>
      {error && <p className="mt-1.5 text-xs text-destructive">{error}</p>}
    </div>
  );
}
