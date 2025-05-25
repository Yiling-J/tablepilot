# Product Catalog Generator Workflow

This workflow automates the creation of a product catalog for a specified category.

## Workflow Steps

1.  **Define Product Category:**
    *   A variable `product_category` is defined to specify the type of products (e.g., "fantasy creatures", "sci-fi gadgets").
    *   The default category is "fantasy creatures".

2.  **Create Product Table:**
    *   A table is created to store product information. The table name is dynamically generated based on the `product_category` (e.g., `fantasy creatures_products`).
    *   The table schema is defined in `products.json` and includes the following columns:
        *   `Name`: (String) The name of the product.
        *   `Description`: (String) An AI-generated description based on the product's name and category.
        *   `Price`: (Number) An AI-generated realistic price for the product.

3.  **Add Image Column:**
    *   An `Image` column (type: image) is added to the product table to store product images.

4.  **Generate Product Entries:**
    *   Three product entries are automatically generated for the specified `product_category`.
    *   The generation process uses a prompt that incorporates the `product_category` to ensure relevant product creation.

5.  **Export Catalog:**
    *   The final product table is exported to a CSV file. The file name is dynamically generated based on the `product_category` (e.g., `fantasy creatures_catalog.csv`).

## Files

*   `workflow.json`: Defines the workflow steps and configurations.
*   `products.json`: Specifies the schema for the product table.
*   `README.md`: This file, providing an overview of the workflow.

## How to Use

1.  Modify the `product_category` variable in `workflow.json` if you want to generate a catalog for a different category.
2.  Run the workflow.
3.  The generated catalog will be available as a CSV file in the workflow's output directory.
