This example demonstrates how to use **Tablepilot** to generate images in three different ways:

1. Generate images together with text columns
2. Autofill an image column based on text columns
3. Autofill an image column based on another image column

It uses Gemini's new experimental image generation model: `gemini-2.0-flash-exp-image-generation`. Currently, this is the **only supported** image generation model.
**Note:** The image generation feature is experimental and may change in future releases.

The full example consists of two tables:
- `recipes`: used in the first example (image + text generation)
- `recipes_v2`: used in the second and third examples (autofill image column)

https://github.com/user-attachments/assets/17c33af4-ed8d-44f9-8141-f18839a2927c



---

### Configuration

Before running any generation commands, update your config file to include the image client and model:

```toml
[common]
source_data_dir = "./"

[database]
driver = "sqlite3"
dsn = "data.db?_pragma=foreign_keys(1)"

[server]
address = ":8080"

[[clients]]
name = "gemini"
type = "openai"
key = "key"
base_url = "https://generativelanguage.googleapis.com/v1beta/openai/"

[[clients]]
name = "gemini-image"
type = "gemini"
key = "key"

[[models]]
model = "gemini-2.0-flash-001"
client = "gemini"
rpm = 10

[[image_models]]
model = "gemini-2.0-flash-exp-image-generation"
client = "gemini-image"
rpm = 10
```

Once configured, Tablepilot will use this image model when generating images.

---

## 1. Generate Images Alongside Text

This is the simplest method: generate image and text together in one pass.

**Step 1: Create the table**
```bash
tablepilot create examples/recipes_with_image/recipes.json
```

This table contains two image columns: one for the recipe and one for ingredients. Both images will be generated along with the text.

**Step 2: Generate data**
```bash
tablepilot generate recipes -c 30 -b 1
```

This sets the batch size to 1, which generates 2 images per API call.

The `gemini-2.0-flash-exp-image-generation` model (free version) is not very stable, and often fails, so it's strongly recommended to use a small batch size (1 or 2).

---

## 2. Autofill Image Column from Text Columns

Because of limitations with the image model, a more robust approach is to generate text first and then fill in the image column. This has several advantages:

- Use large batches when generating text (fast and reliable)
- Autofill images in smaller batches for better stability
- Resume image generation using the `--offset` flag if an error occurs

**Step 1: Create the table (no image columns yet)**
```bash
tablepilot create examples/recipes_with_image/recipes_v2.json
```

**Step 2: Generate text data**
```bash
tablepilot generate recipes_v2 -c 30 -b 6
```

**Step 3: Add an image column**
```bash
tablepilot update examples/recipes_with_image/recipes_v2_with_image.json
```

**Step 4: Autofill the image column using text context**
```bash
tablepilot autofill recipes_v2 -c 30 -b 2 \
  --columns=Image \
  --context_columns=Name \
  --context_columns=Ingredients \
  --context_columns=Country \
  --context_columns=Steps
```

To resume from a specific offset after a failure:
```bash
tablepilot autofill recipes_v2 -c 30 -b 2 \
  --columns=Image \
  --context_columns=Name \
  --context_columns=Ingredients \
  --context_columns=Country \
  --context_columns=Steps \
  -o=20
```

---

## 3. Autofill Image Column from Another Image Column

This example generates a new image based on an existing image column. The new image adds a drink to the original recipe photo, demonstrating image-to-image generation capabilities.

**Step 1: Update the table to include the new image column**
```bash
tablepilot update examples/recipes_with_image/recipes_v2_with_image_combo.json
```

**Step 2: Autofill the new column using an existing image column**
```bash
tablepilot autofill recipes_v2 -c 30 -b 2 \
  --columns="Combo Meal" \
  --context_columns=Image
```
