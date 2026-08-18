export const moods = [
  { id: "neutral", label: "Нейтрально" },
  { id: "happy", label: "Радостно" },
  { id: "sad", label: "Грустно" },
  { id: "ironic", label: "Иронично" }
];

export const defaultMood = "neutral";

export function getMoodLabel(mood) {
  return moods.find((item) => item.id === mood)?.label || mood;
}
