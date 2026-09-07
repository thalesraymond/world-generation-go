const regex = /^([a-z]+)(\([^)]+\))?!?: /;
const titles = [
  "⚡ Bolt: Optimize demographic automata iteration loops",
  "perf: ⚡ Bolt: Optimize demographic automata iteration loops",
  "perf: optimize demographic automata iteration loops",
];
titles.forEach(title => {
  console.log(title, regex.test(title));
});
