<script>
  let { onGenerating } = $props();

  let content = $state('');
  let language = $state('en');
  let durationTarget = $state('medium');
  let tone = $state('casual');
  let persona = $state('');
  let topicHint = $state('');
  let previewOnly = $state(false);

  const languages = [
    { code: 'en', label: 'English' },
    { code: 'es', label: 'Español' },
    { code: 'fr', label: 'Français' },
    { code: 'de', label: 'Deutsch' },
    { code: 'pt', label: 'Português' },
    { code: 'it', label: 'Italiano' },
  ];

  const durations = [
    { value: 'short', label: 'Short (~2 min)', words: '~300 words' },
    { value: 'medium', label: 'Medium (~5 min)', words: '~700 words' },
    { value: 'long', label: 'Long (~10 min)', words: '~1400 words' },
  ];

  const tones = [
    { value: 'casual', label: 'Casual', desc: 'Friendly, approachable' },
    { value: 'academic', label: 'Academic', desc: 'Formal, structured' },
    { value: 'enthusiastic', label: 'Enthusiastic', desc: 'Energetic, engaging' },
    { value: 'calm', label: 'Calm', desc: 'Relaxed, methodical' },
  ];

  async function handleSubmit(e) {
    e.preventDefault();
    if (!content.trim()) return;
    onGenerating();
    // Tool call will be handled by the host
  }
</script>

<form class="podcast-form" onsubmit={handleSubmit}>
  <div class="field">
    <label for="content">Study Material</label>
    <textarea
      id="content"
      bind:value={content}
      placeholder="Paste your notes, textbook excerpt, or study material here..."
      rows="5"
      required
    ></textarea>
  </div>

  <div class="row">
    <div class="field">
      <label for="language">Language</label>
      <select id="language" bind:value={language}>
        {#each languages as lang}
          <option value={lang.code}>{lang.label}</option>
        {/each}
      </select>
    </div>

    <div class="field">
      <label for="duration">Duration</label>
      <select id="duration" bind:value={durationTarget}>
        {#each durations as d}
          <option value={d.value}>{d.label}</option>
        {/each}
      </select>
    </div>
  </div>

  <div class="field">
    <label>Tone</label>
    <div class="tone-grid">
      {#each tones as t}
        <button
          type="button"
          class="tone-btn"
          class:active={tone === t.value}
          onclick={() => tone = t.value}
        >
          <span class="tone-label">{t.label}</span>
          <span class="tone-desc">{t.desc}</span>
        </button>
      {/each}
    </div>
  </div>

  <div class="field">
    <label for="topic">Topic Hint <span class="optional">(optional)</span></label>
    <input
      id="topic"
      type="text"
      bind:value={topicHint}
      placeholder="e.g., Krebs Cycle, Spanish Revolution"
    />
  </div>

  <div class="field">
    <label for="persona">Custom Persona <span class="optional">(optional)</span></label>
    <input
      id="persona"
      type="text"
      bind:value={persona}
      placeholder="e.g., a friendly tutor, a history professor"
    />
  </div>

  <div class="actions">
    <label class="preview-toggle">
      <input type="checkbox" bind:checked={previewOnly} />
      <span>Preview script only</span>
    </label>
    <button type="submit" class="btn-primary" disabled={!content.trim()}>
      {previewOnly ? 'Preview Script' : 'Generate Podcast'}
    </button>
  </div>
</form>

<style>
  .podcast-form {
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  label {
    font-size: 12px;
    font-weight: 600;
    color: #4a4a4a;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .optional {
    font-weight: 400;
    text-transform: none;
    letter-spacing: 0;
    color: #999;
  }

  textarea, input, select {
    padding: 12px 14px;
    border: 1px solid #e8e5e0;
    border-radius: 10px;
    font-size: 14px;
    font-family: inherit;
    background: #fff;
    color: #2d2d2d;
    transition: border-color 0.2s, box-shadow 0.2s;
  }

  textarea:focus, input:focus, select:focus {
    outline: none;
    border-color: #8b7355;
    box-shadow: 0 0 0 3px rgba(139, 115, 85, 0.1);
  }

  textarea {
    resize: vertical;
    min-height: 100px;
  }

  textarea::placeholder, input::placeholder {
    color: #b0a89e;
  }

  .row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  select {
    cursor: pointer;
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%236b6b6b' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 12px center;
    padding-right: 32px;
  }

  .tone-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
  }

  .tone-btn {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    padding: 10px 12px;
    border: 1px solid #e8e5e0;
    border-radius: 10px;
    background: #fff;
    cursor: pointer;
    transition: all 0.2s;
    text-align: left;
  }

  .tone-btn:hover {
    border-color: #d0cbc3;
    background: #faf8f5;
  }

  .tone-btn.active {
    border-color: #8b7355;
    background: #f9f5ee;
  }

  .tone-label {
    font-size: 13px;
    font-weight: 600;
    color: #2d2d2d;
  }

  .tone-desc {
    font-size: 11px;
    color: #8a8a8a;
  }

  .actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding-top: 4px;
  }

  .preview-toggle {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: #6b6b6b;
    cursor: pointer;
    text-transform: none;
    letter-spacing: 0;
    font-weight: 400;
  }

  .preview-toggle input {
    width: 16px;
    height: 16px;
    accent-color: #8b7355;
  }

  .btn-primary {
    padding: 12px 24px;
    background: linear-gradient(135deg, #8b7355 0%, #a08060 100%);
    color: #fff;
    border: none;
    border-radius: 10px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 2px 8px rgba(139, 115, 85, 0.25);
  }

  .btn-primary:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(139, 115, 85, 0.35);
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
