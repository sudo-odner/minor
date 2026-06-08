import { useState, useEffect } from 'react';

/**
 * Хук для задержки обновления значения (debouncing).
 * Помогает избежать слишком частых вызовов API при вводе текста.
 */
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
}
