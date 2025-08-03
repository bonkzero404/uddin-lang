/**
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
import siteConfig from '@generated/docusaurus.config';
import ExecutionEnvironment from '@docusaurus/ExecutionEnvironment';

export default function prismIncludeLanguages(PrismObject) {
  const {
    themeConfig: {prism},
  } = siteConfig;
  const {additionalLanguages} = prism;
  
  // Prism components work on the Prism instance on the window, while prism-
  // react-renderer uses its own Prism instance. We temporarily mount the
  // instance onto window, import components to enhance it, then remove it to
  // avoid polluting global namespace.
  // You can mutate PrismObject: registering plugins, deleting languages... As
  // long as you don't re-assign it
  globalThis.Prism = PrismObject;
  
  additionalLanguages.forEach((lang) => {
    if (lang === 'uddin') {
      // Define Uddin-Lang syntax for Prism
      PrismObject.languages.uddin = {
        'comment': {
          pattern: /\/\/.*$/m,
          greedy: true
        },
        'string': {
          pattern: /"(?:[^"\\\r\n]|\\.)*"/,
          greedy: true
        },
        'keyword': /\b(?:func|end|if|then|else|elif|while|for|do|break|continue|return|try|catch|import|as|true|false|null)\b/,
        'builtin': /\b(?:print|println|len|str|split|join|trim|replace|substring|indexOf|toLowerCase|toUpperCase|startsWith|endsWith|abs|ceil|floor|round|sqrt|pow|sin|cos|tan|log|exp|min|max|random|append|filter|map|reduce|sort|reverse|slice|contains|indexOf|lastIndexOf|date_now|date_format|date_parse|date_format_new|date_add|date_subtract|date_diff|date_between|date_compare|regex_match|regex_find|regex_find_all|regex_replace|regex_split|is_regex_match|http_get|http_post|http_put|http_delete|http_patch|tcp_connect|tcp_listen|tcp_close|udp_send|udp_receive|read_file|write_file|append_file|file_exists|file_size|file_delete|create_directory|list_directory|file_copy|file_move)\b/,
        'number': /\b\d+(?:\.\d+)?\b/,
        'operator': /[+\-*/%=<>!&|]+/,
        'punctuation': /[{}\[\]();,.:]/,
        'variable': /\b[a-zA-Z_][a-zA-Z0-9_]*\b/
      };
      
      // Also register 'din' as an alias for 'uddin'
      PrismObject.languages.din = PrismObject.languages.uddin;
    } else {
      // eslint-disable-next-line global-require, import/no-dynamic-require
      require(`prismjs/components/prism-${lang}`);
    }
  });
  
  delete globalThis.Prism;
}