# Documentation Images

This directory contains images and media files used in the documentation site.

## Required Branding Assets

### Logo Files
- `deepscanbot-logo.png` - Main logo for navbar and favicon (recommended: 512x512px, PNG with transparency)
- `deepscanbot-social-preview.png` - Social media preview image (recommended: 1200x630px, PNG or JPG)

### Scan Result Screenshots
Add the following screenshots to demonstrate DeepScanBot in action:

- `scan-terminal-example.png` - Screenshot of terminal showing a scan in progress
- `scan-text-output.png` - Screenshot showing crawler_results.txt file contents
- `scan-json-output.png` - Screenshot showing crawler_results.json file contents
- `scan-demo.gif` - Optional: Short GIF demonstrating a scan from start to finish

## Image Guidelines

- Use PNG format for logos and screenshots with text
- Use JPG for photos or complex images
- Keep file sizes under 500KB for optimal loading
- Use descriptive filenames with hyphens

## Adding Images to Documentation

Reference images in MDX files using site-root paths (beginning with /img/):

```markdown
![Scan Terminal Example](/img/scan-terminal-example.png)
```

Or in Docusaurus tabs/cards:

```jsx
<Tabs>
  <TabItem value="text" label="Text Output">
    ![Text Output Example](/img/scan-text-output.png)
  </TabItem>
</Tabs>
```
