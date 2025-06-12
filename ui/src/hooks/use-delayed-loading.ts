import { useState, useEffect, useRef } from 'react';

export function useDelayedLoading(
  actualLoading: boolean,
  minDelay: number = 300 // Default minimum delay of 300ms
): boolean {
  const [delayedLoading, setDelayedLoading] = useState(actualLoading);
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    if (actualLoading) {
      setDelayedLoading(true);
      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
    } else {
      // If actual loading is false, we want to keep delayedLoading true
      // until minDelay has passed since actualLoading became false.
      timerRef.current = setTimeout(() => {
        setDelayedLoading(false);
        timerRef.current = null;
      }, minDelay);
    }

    // Cleanup timeout on unmount or if actualLoading changes back to true
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
    };
  }, [actualLoading, minDelay]);

  // Ensure initial state is correct
  useEffect(() => {
    setDelayedLoading(actualLoading);
  }, [actualLoading]);

  return delayedLoading;
}
