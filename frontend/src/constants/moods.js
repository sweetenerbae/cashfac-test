export const moods = [
  { id: "neutral", label: "Нейтрально", emoji: "📰" },
  { id: "happy", label: "Радостно", emoji: "✨" },
  { id: "sad", label: "Грустно", emoji: "☔" },
  { id: "ironic", label: "Иронично", emoji: "🎭" }
];

export const defaultMood = "neutral";

export function getMoodLabel(mood) {
  return moods.find((item) => item.id === mood)?.label || mood;
}
