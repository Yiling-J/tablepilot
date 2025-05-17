import { JSONObject } from "@/json";
import { imageUrl } from "@/urls.tsx";

export function DataCard({ data }: { data: JSONObject }) {
  return (
    <div className="bg-gray-800 border border-gray-700 rounded-md p-3 my-2">
      <div className="grid grid-cols-1">
        {Object.entries(data).map(([key, value]) => (
          <div key={key} className="flex flex-col mb-1">
            <span className="text-xs text-gray-500 mb-1">{key}</span>
            {(value as JSONObject).type === "image" ? (
              <div>
                <img
                  src={imageUrl((value as JSONObject).value as string)}
                  width={256}
                  height={256}
                  className="rounded"
                />
              </div>
            ) : (
              <span className="text-sm text-white font-medium">
                {(value as JSONObject).value as string}
              </span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
