<script>
  let { runTool } = $props();

  let content = $state('');
  let language = $state('en');
  let difficulty = $state('graduate');
  let durationTarget = $state('medium');
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
    { value: 'short', label: 'Short (~5 min)', words: '~2,000 words' },
    { value: 'medium', label: 'Medium (~15 min)', words: '~4,500 words' },
    { value: 'long', label: 'Long (~30 min)', words: '~8,000 words' },
  ];

  const difficulties = [
    { value: 'undergraduate', label: 'Undergraduate', desc: 'Foundational' },
    { value: 'graduate', label: 'Graduate', desc: 'Comprehensive' },
    { value: 'masters', label: "Master's", desc: 'Expert depth' },
    { value: 'phd', label: 'PhD', desc: 'Full research level' },
  ];

  async function handleSubmit(e) {
    e.preventDefault();
    if (!content.trim()) return;
    await runTool('generate_study_guide', {
      content,
      language,
      difficulty,
      durationTarget,
      previewOnly,
    });
  }
</script>

<form class="guide-form" onsubmit={handleSubmit}>
  <div class="field">
    <label for="content">Study Material</label>
    <textarea
      id="content"
      bind:value={content}
      placeholder="Paste your notes, textbook excerpt, or study material to generate a deep, expert-level study guide..."
      rows="6"
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
    <span id="difficulty-label" class="field-label">Academic Level</span>
    <div class="difficulty-grid" role="radiogroup" aria-labelledby="difficulty-label">
      {#each difficulties as d}
        <button
          type="button"
          class="difficulty-btn"
          class:active={difficulty === d.value}
          role="radio"
          aria-checked={difficulty === d.value}
          onclick={() => difficulty = d.value}
        >
          <span class="diff-label">{d.label}</span>
          <span class="diff-desc">{d.desc}</span>
        </button>
      {/each}
    </div>
  </div>

  <div class="actions">
    <label class="preview-toggle">
      <input type="checkbox" bind:checked={previewOnly} />
      <span>Preview script only</span>
    </label>
    <button type="submit" class="btn-primary" disabled={!content.trim()}>
      {previewOnly ? 'Preview Script' : 'Generate Study Guide'}
    </button>
  </div>
</form>

<style>
  .guide-form {
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  label, .field-label {
    font-size: 12px;
    font-weight: 600;
    color: #4a4a4a;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  textarea, select {
    padding: 12px 14px;
    border: 1px solid #e8e5e0;
    border-radius: 10px;
    font-size: 14px;
    font-family: inherit;
    background: #fff;
    color: #2d2d2d;
    transition: border-color 0.2s, box-shadow 0.2s;
  }

  textarea:focus, select:focus {
    outline: none;
    border-color: #8b7355;
    box-shadow: 0 0 0 3px rgba(139, 115, 85, 0.1);
  }

  textarea {
    resize: vertical;
    min-height: 120px;
  }

  textarea::placeholder {
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

  .difficulty-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
  }

  .difficulty-btn {
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

  .difficulty-btn:hover {
    border-color: #d0cbc3;
    background: #faf8f5;
  }

  .difficulty-btn.active {
    border-color: #8b7355;
    background: #f9f5ee;
  }

  .diff-label {
    font-size: 13px;
    font-weight: 600;
    color: #2d2d2d;
  }

  .diff-desc {
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
    background: linear-gradient(135deg, #6b5b7a 0%, #8b7099 100%);
    color: #fff;
    border: none;
    border-radius: 10px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 2px 8px rgba(107, 91, 122, 0.25);
  }

  .btn-primary:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(107, 91, 122, 0.35);
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
