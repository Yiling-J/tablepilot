import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import { customAlphabet } from "nanoid";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export const nanoid = customAlphabet(
  "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
  7,
); // 7-character random string

const chunkifyRegexp = /(\d+\.\d+|\d+|\D+)/g; // Regular expression for numbers including floats

function chunkify(s: string): string[] {
  return s.match(chunkifyRegexp) || [];
}

export function compareNature(a: string, b: string): number {
  // Handle empty strings: empty string is always smaller
  if (a === "") return b === "" ? 0 : -1;
  if (b === "") return 1;

  const chunksA = chunkify(a);
  const chunksB = chunkify(b);

  const nChunksA = chunksA.length;
  const nChunksB = chunksB.length;

  for (let i = 0; i < chunksA.length; i++) {
    if (i >= nChunksB) {
      return -1; // a comes before b
    }

    const aNum = Number(chunksA[i]);
    const bNum = Number(chunksB[i]);

    // If both chunks are numeric, compare them as integers
    if (!isNaN(aNum) && !isNaN(bNum)) {
      if (aNum !== bNum) {
        return aNum - bNum; // Return the difference for sorting
      }

      // If numbers are equal, continue to the next chunk
      continue;
    }

    // Compare as strings if either chunk is not a number
    if (chunksA[i] !== chunksB[i]) {
      return chunksA[i] < chunksB[i] ? -1 : 1;
    }

    // If both chunks are equal, continue to the next chunk
  }

  // If all chunks are equal up to this point
  return nChunksA - nChunksB; // Shorter string comes first
}

export function naturalSort(l: string[]): string[] {
  return [...l].sort(compareNature);
}
