import { cn } from "@/lib/utils";
import type React from "react";
import { useEffect, useRef, useState, type KeyboardEvent } from "react";
import { Input, InputProps } from "./input";
import { Textarea, TextareaProps } from "./textarea";

export type ContextVariable = {
  display: string;
  path: string;
  type: string;
};

type MentionInputProps = (InputProps & TextareaProps) & {
  variables?: ContextVariable[];
  textarea?: boolean;
  rows?: number;
};

export function MentionInput({
  value,
  variables,
  onChange,
  placeholder = "Type @ to mention a variable...",
  className,
  textarea = false,
  rows = 1,
  ...rest
}: MentionInputProps) {
  if (variables === undefined) {
    if (textarea) {
      return (
        <Textarea
          className={className}
          rows={rows}
          value={value}
          placeholder={placeholder}
          onChange={onChange}
          {...rest}
        />
      );
    } else {
      return (
        <Input
          className={className}
          value={value}
          placeholder={placeholder}
          onChange={onChange}
          {...rest}
        />
      );
    }
  }

  const [showDropdown, setShowDropdown] = useState(false);
  const [filteredVars, setFilteredVars] = useState<ContextVariable[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);

  const inputRef = useRef<HTMLDivElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Handle input changes
  const handleInput = (e: React.FormEvent<HTMLDivElement>) => {
    // Build plain text from HTML nodes to preserve line breaks
    let content = "";
    for (let i = 0; i < e.currentTarget.childNodes.length; i++) {
      const node = e.currentTarget.childNodes[i];
      if (node.nodeType === Node.ELEMENT_NODE) {
        if (node.nodeName === "SPAN") {
        }
        if (node.nodeName === "BR") {
          content += "\n";
        } else if (node.nodeName === "DIV") {
          if (i > 0) content += "\n"; // Add newline before div except for first div
          content += node.textContent || "";
        } else {
          content += node.textContent || "";
        }
      } else if (node.nodeType === Node.TEXT_NODE) {
        content += node.textContent || "";
      }
    }

    // Check if @ was just typed
    if (content.includes("@") && !showDropdown) {
      setShowDropdown(true);
      setFilteredVars(variables);
      setSelectedIndex(0);
    }

    // Update search text if dropdown is open
    if (showDropdown) {
      const atIndex = content.lastIndexOf("@");
      const searchStr = content.substring(atIndex + 1);

      // Filter variables based on search text
      const filtered = variables.filter((v) =>
        v.display.toLowerCase().includes(searchStr.toLowerCase()),
      );
      setFilteredVars(filtered);
      setSelectedIndex(0);
    }

    // Notify parent component
    if (onChange) {
      const formatted = formatInputWithVariables(content);
      const event = {
        target: { value: formatted },
      } as React.ChangeEvent<HTMLInputElement>;
      onChange(event);
    }
  };

  // Handle key presses
  const handleKeyDown = (e: KeyboardEvent<HTMLDivElement>) => {
    if (showDropdown) {
      switch (e.key) {
        case "ArrowDown":
          e.preventDefault();
          setSelectedIndex((prev) => (prev + 1) % filteredVars.length);
          break;
        case "ArrowUp":
          e.preventDefault();
          setSelectedIndex(
            (prev) => (prev - 1 + filteredVars.length) % filteredVars.length,
          );
          break;
        case "Enter":
        case "Tab":
          e.preventDefault();
          if (filteredVars.length > 0) {
            insertMention(filteredVars[selectedIndex]);
          }
          break;
        case "Escape":
          e.preventDefault();
          setShowDropdown(false);
          break;
      }
    }
  };

  const buildElements = (raw: string) => {
    if (!inputRef.current) return;

    // Clear existing content
    while (inputRef.current.firstChild) {
      inputRef.current.removeChild(inputRef.current.firstChild);
    }

    // Split the input by newlines and process each line
    const lines = raw.split("\n");
    const mentionRegex = /{{\.(.*?)}}/g;

    lines.forEach((line, lineIndex) => {
      if (!inputRef.current) return;

      let lastIndex = 0;
      let match;

      while ((match = mentionRegex.exec(line)) !== null) {
        if (!inputRef.current) return;

        const [fullMatch, path] = match;

        // Add text before the mention
        if (match.index > lastIndex) {
          const textBefore = line.substring(lastIndex, match.index);
          inputRef.current.appendChild(document.createTextNode(textBefore));
        }

        // Add the mention
        const variable = variables?.find((v) => v.path === path);
        if (variable) {
          const mentionSpan = document.createElement("span");
          mentionSpan.className =
            "inline-block bg-primary border border-gray-300 rounded px-1 mx-0.5 text-secondary";
          mentionSpan.contentEditable = "false";
          mentionSpan.dataset.path = variable.path;
          mentionSpan.textContent = variable.display;
          inputRef.current.appendChild(mentionSpan);
        } else {
          inputRef.current.appendChild(document.createTextNode(fullMatch));
        }

        lastIndex = mentionRegex.lastIndex;
      }

      // Add remaining text in the line
      if (lastIndex < line.length) {
        if (!inputRef.current) return;
        const textAfter = line.substring(lastIndex);
        inputRef.current.appendChild(document.createTextNode(textAfter));
      }

      // Add newline after each line except the last one
      if (lineIndex < lines.length - 1) {
        if (!inputRef.current) return;
        const br = document.createElement("br");
        inputRef.current.appendChild(br);
      }
    });

    // Update plain text state
    if (!inputRef.current) return;
  };

  const insertMention = (variable: ContextVariable) => {
    if (!inputRef.current) return;

    // Get the current selection
    const selection = window.getSelection();
    const range = selection?.getRangeAt(0);
    if (!range) return;

    // Get the current content as HTML
    const currentHtml = inputRef.current.innerHTML;

    // Find the @ symbol and its position
    const atIndex = currentHtml.lastIndexOf("@");
    if (atIndex === -1) return;

    // Split the HTML at the @ symbol
    const beforeAt = currentHtml.substring(0, atIndex);
    const afterAt = currentHtml.substring(atIndex + 1);

    // Create the mention span
    const mentionSpan = document.createElement("span");
    let vid = document.createAttribute("vid");
    const vidv = Math.random().toString(20);
    vid.value = vidv;
    mentionSpan.className =
      "inline-block bg-primary border border-gray-300 rounded px-1 mx-0.5 text-secondary";
    mentionSpan.contentEditable = "false";
    mentionSpan.dataset.path = variable.path;
    mentionSpan.textContent = variable.display;
    mentionSpan.attributes.setNamedItem(vid);

    // Create a temporary div to get the HTML of the mention span
    const tempDiv = document.createElement("div");
    tempDiv.appendChild(mentionSpan);
    const mentionHtml = tempDiv.innerHTML;

    let newHtml = beforeAt + mentionHtml;
    // Combine the parts with the mention
    if (afterAt.length > 0) {
      newHtml = newHtml + " " + afterAt;
    }

    // Update the content
    inputRef.current.innerHTML = newHtml;

    // Close the dropdown
    setShowDropdown(false);

    // Set focus back to the input
    inputRef.current.focus();

    // Position cursor at the end
    const newSelection = window.getSelection();
    const newRange = document.createRange();
    const el = inputRef.current.querySelector(`span[vid='${vidv}']`);

    if (el) {
      newRange.setStartAfter(el);
      newRange.setEndAfter(el);
      newSelection?.removeAllRanges();
      newSelection?.addRange(newRange);
    }

    // Notify parent component
    if (onChange) {
      const formatted = formatInputWithVariables(
        inputRef.current.textContent || "",
      );
      const event = {
        target: { value: formatted },
      } as React.ChangeEvent<HTMLInputElement>;
      onChange(event);
    }
  };

  // Format input by replacing variable mentions with {{.var}} format
  const formatInputWithVariables = (input: string): string => {
    if (!inputRef.current) return input;

    let result = "";

    // Go through all child nodes
    for (let i = 0; i < inputRef.current.childNodes.length; i++) {
      const node = inputRef.current.childNodes[i];
      if (node.nodeType === Node.ELEMENT_NODE) {
        if (node.nodeName === "SPAN") {
          const span = node as HTMLSpanElement;
          if (span.dataset.path) {
            // This is a mention tag, format it as {{.path}}
            result += `{{.${span.dataset.path}}}`;
          } else {
            // Regular span
            result += node.textContent || "";
          }
        } else if (node.nodeName === "BR") {
          result += "\n";
        } else if (node.nodeName === "DIV") {
          if (i > 0) result += "\n"; // Add newline before div except for first div
          for (let i = 0; i < node.childNodes.length; i++) {
            const inode = node.childNodes[i];
            if (inode.nodeName === "SPAN") {
              const span = inode as HTMLSpanElement;
              if (span.dataset.path) {
                // This is a mention tag, format it as {{.path}}
                result += `{{.${span.dataset.path}}}`;
              } else {
                // Regular span
                result += inode.textContent || "";
              }
            } else {
              result += inode.textContent || "";
            }
          }
        } else {
          result += node.textContent || "";
        }
      } else if (node.nodeType === Node.TEXT_NODE) {
        result += node.textContent || "";
      }
    }

    return result;
  };

  // Handle clicking on a variable in the dropdown
  const handleVariableClick = (variable: ContextVariable) => {
    insertMention(variable);
  };

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(e.target as Node) &&
        inputRef.current &&
        !inputRef.current.contains(e.target as Node)
      ) {
        setShowDropdown(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  useEffect(() => {
    buildElements(value as string);
  }, []);

  const minHeightVariants = {
    1: "min-h-[30px]",
    2: "min-h-[60px]",
    3: "min-h-[90px]",
    4: "min-h-[120px]",
    5: "min-h-[150px]",
  } as Record<number, string>;

  return (
    <div
      className={cn(
        "relative",
        textarea ? (minHeightVariants[rows] ?? "min-h-[80px]") : "min-h-[40px]",
      )}
    >
      <div
        ref={inputRef}
        contentEditable
        className={cn(
          "w-full rounded-md border border-input px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
          textarea
            ? (minHeightVariants[rows] ?? "min-h-[80px]")
            : "min-h-[40px]",
          className,
        )}
        onInput={handleInput}
        onKeyDown={handleKeyDown}
        data-placeholder={placeholder}
        onFocus={() => {
          if (!inputRef.current?.textContent) {
            inputRef.current?.setAttribute("data-empty", "true");
          } else {
            inputRef.current?.removeAttribute("data-empty");
          }
        }}
        onBlur={() => {
          inputRef.current?.removeAttribute("data-empty");
        }}
      />

      {/* Placeholder text */}
      {!value && (
        <div className="absolute top-0 left-0 px-3 py-2 text-sm text-muted-foreground pointer-events-none">
          {placeholder}
        </div>
      )}

      {/* Dropdown for variables */}
      {showDropdown && (
        <div
          ref={dropdownRef}
          className="absolute z-10 mt-1 w-full max-h-60 overflow-auto rounded-md border border-gray-200 shadow-lg bg-background"
        >
          {filteredVars.length > 0 ? (
            <ul className="py-1">
              {filteredVars.map((variable, index) => (
                <li
                  key={index}
                  className={cn(
                    "px-3 py-2 text-sm cursor-pointer hover:bg-primary/10",
                    index === selectedIndex ? "bg-primary/10" : "",
                  )}
                  onClick={() => handleVariableClick(variable)}
                >
                  <span className="font-medium">{variable.display}</span>
                </li>
              ))}
            </ul>
          ) : (
            <div className="px-3 py-2 text-sm text-gray-500">
              No variables found
            </div>
          )}
        </div>
      )}
    </div>
  );
}
