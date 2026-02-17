interface SearchHighlightProps {
  text: string;
  query: string;
}

function escapeRegex(str: string): string {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

export function SearchHighlight({ text, query }: SearchHighlightProps) {
  if (!query || query.length < 2) {
    return <>{text}</>;
  }

  const escaped = escapeRegex(query);
  const parts = text.split(new RegExp(`(${escaped})`, 'gi'));

  return (
    <>
      {parts.map((part, i) => {
        const isMatch = part.toLowerCase() === query.toLowerCase();
        const key = `${part}-${String(i)}`;
        if (isMatch) {
          return (
            <mark key={key} className="bg-primary/30 text-foreground rounded-sm">
              {part}
            </mark>
          );
        }
        return <span key={key}>{part}</span>;
      })}
    </>
  );
}
