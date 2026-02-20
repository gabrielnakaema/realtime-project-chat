import { $isTextNode, ParagraphNode, TextNode, isHTMLElement } from 'lexical';

import type { DOMConversionMap, DOMExportOutput, DOMExportOutputMap, Klass, LexicalEditor, LexicalNode } from 'lexical';

export const removeStylesExportDOM = (editor: LexicalEditor, target: LexicalNode): DOMExportOutput => {
  const output = target.exportDOM(editor);
  if (isHTMLElement(output.element)) {
    for (const el of [output.element, ...output.element.querySelectorAll('[style],[class]')]) {
      el.removeAttribute('class');
      el.removeAttribute('style');
    }
  }
  return output;
};

export const exportParagraphDOM = (editor: LexicalEditor, target: LexicalNode): DOMExportOutput => {
  const output = target.exportDOM(editor);
  if (isHTMLElement(output.element)) {
    const textAlign = output.element.style.textAlign;
    for (const el of [output.element, ...output.element.querySelectorAll('[style],[class]')]) {
      el.removeAttribute('class');
      el.removeAttribute('style');
    }
    if (textAlign) {
      output.element.style.textAlign = textAlign;
    }
  }
  return output;
};

export const exportMap: DOMExportOutputMap = new Map<
  Klass<LexicalNode>,
  (editor: LexicalEditor, target: LexicalNode) => DOMExportOutput
>([
  [ParagraphNode, exportParagraphDOM],
  [TextNode, removeStylesExportDOM],
]);

export const getExtraStyles = (element: HTMLElement): string => {
  let extraStyles = '';
  const fontSize = parseAllowedFontSize(element.style.fontSize);
  const backgroundColor = parseAllowedColor(element.style.backgroundColor);
  const color = parseAllowedColor(element.style.color);
  if (fontSize !== '' && fontSize !== '15px') {
    extraStyles += `font-size: ${fontSize};`;
  }
  if (backgroundColor !== '' && backgroundColor !== 'rgb(255, 255, 255)') {
    extraStyles += `background-color: ${backgroundColor};`;
  }
  if (color !== '' && color !== 'rgb(0, 0, 0)') {
    extraStyles += `color: ${color};`;
  }
  return extraStyles;
};

export const constructImportMap = (): DOMConversionMap => {
  const importMap: DOMConversionMap = {};

  for (const [tag, fn] of Object.entries(TextNode.importDOM() || {})) {
    importMap[tag] = (importNode) => {
      const importer = fn(importNode);
      if (!importer) {
        return null;
      }
      return {
        ...importer,
        conversion: (element) => {
          const output = importer.conversion(element);
          if (output === null || output.forChild === undefined || output.after !== undefined || output.node !== null) {
            return output;
          }
          const extraStyles = getExtraStyles(element);
          if (extraStyles) {
            const { forChild } = output;
            return {
              ...output,
              forChild: (child, parent) => {
                const textNode = forChild(child, parent);
                if ($isTextNode(textNode)) {
                  textNode.setStyle(textNode.getStyle() + extraStyles);
                }
                return textNode;
              },
            };
          }
          return output;
        },
      };
    };
  }

  return importMap;
};

const MIN_ALLOWED_FONT_SIZE = 8;
const MAX_ALLOWED_FONT_SIZE = 72;

export const parseAllowedFontSize = (input: string): string => {
  const match = input.match(/^(\d+(?:\.\d+)?)px$/);
  if (match) {
    const n = Number(match[1]);
    if (n >= MIN_ALLOWED_FONT_SIZE && n <= MAX_ALLOWED_FONT_SIZE) {
      return input;
    }
  }
  return '';
};

export const parseAllowedColor = (input: string) => {
  return /^rgb\(\d+, \d+, \d+\)$/.test(input) ? input : '';
};
