/**
 * @module json
 * @description
 *
 * This module provides types and interfaces to ensure type-safety
 * when working with JSON structures in TypeScript. It ensures that
 * values, arrays, and objects conform to the JSON format, preventing
 * the use of non-JSON-safe values.
 */

/**
 * Represents a valid JSON value. In JSON, values must be one of:
 * string, number, boolean, null, array or object.
 *
 * @type JSONValue
 */
export type JSONValue =
  | string
  | number
  | boolean
  | null
  | JSONArray
  | JSONObject;

/**
 * Represents a valid JSON array. A JSON array is a list of `JSONValue` items.
 * This extends the built-in Array type to ensure that all its members are
 * valid JSON values.
 *
 * @interface JSONArray
 */
export type JSONArray = Array<JSONValue>;

/**
 * Represents a valid JSON object. A JSON object is a collection of key-value
 * pairs, where each key is a string and each value is a `JSONValue`.
 *
 * @interface JSONObject
 */
export interface JSONObject {
  [key: string]: JSONValue;
}
