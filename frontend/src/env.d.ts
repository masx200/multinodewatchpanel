/// <reference types="vite/client" />

declare module '*.vue' {
    import type { DefineComponent } from 'vue';
    const component: DefineComponent<{}, {}, any>;
    export default component;
}

declare module 'monaco-editor/esm/vs/language/json/monaco.contribution' {}
declare module 'monaco-editor/esm/vs/language/css/monaco.contribution' {}
declare module 'monaco-editor/esm/vs/language/html/monaco.contribution' {}
declare module 'monaco-editor/esm/vs/language/typescript/monaco.contribution' {}
declare module 'monaco-editor/esm/vs/basic-languages/_.contribution.js' {}
declare module 'monaco-editor/esm/vs/editor/contrib/folding/browser/folding.js' {}
declare module 'monaco-editor/esm/vs/editor/contrib/contextmenu/browser/contextmenu.js' {}
declare module 'monaco-editor/esm/vs/editor/contrib/clipboard/browser/clipboard.js' {}
declare module 'monaco-editor/esm/vs/editor/contrib/dropOrPasteInto/browser/copyPasteContribution.js' {}
declare module 'monaco-editor/esm/vs/editor/contrib/find/browser/findController.js' {}
declare module 'monaco-editor/esm/vs/editor/contrib/multicursor/browser/multicursor.js' {}
declare module 'monaco-editor/esm/vs/editor/standalone/browser/quickAccess/standaloneCommandsQuickAccess.js' {}
declare module 'monaco-editor/esm/vs/editor/editor.api' {
    export * from 'monaco-editor';
}
