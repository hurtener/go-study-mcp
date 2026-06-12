<script>
  let { onGenerating } = $props();

  let mode = $state('content');
  let content = $state('');
  let cards = $state([{ question: '', answer: '' }]);
  let language = $state('en');
  let difficulty = $state('intermediate');
  let cardCount = $state(10);
  let pauseDuration = $state(5);
  let previewOnly = $state(false);

  const difficulties = [
    { value: 'basic', label: 'Basic', desc: 'Definition & recall' },
    { value: 'intermediate', label: 'Intermediate', desc: 'Comprehension' },
    { value: 'advanced', label: 'Advanced', desc: 'Application & synthesis' },
  ];

  function addCard() {
    cards = [...cards, { question: '', answer: '' }];
  }

  function removeCard(index) {
    cards = cards.filter((_, i) => i !== index);
  }

  function updateCard(index, field, value) {
    cards = cards.map((c, i) => i === index ? { ...c, [field]: value } : c);
  }

  async function handleSubmit(e) {
    e.preventDefault();
    onGenerating();
  }
</script>

<form class="flashcard-form" onsubmit={handleSubmit}>
  <div class="mode-toggle">
    <button
      type="button"
      class="mode-btn"
      class:active={mode === 'content'}
      onclick={() => mode = 'content'}
    >
      From Material
    </button>
    <button
      type="button"
      class="mode-btn"
      class:active={mode === 'cards'}
      onclick={() => mode = 'cards'}
    >
      My Cards
    </button>
  </div>

  {#if mode === 'content'}
    <div class="field">
      <label for="fc-content">Study Material</label>
      <textarea
        id="fc-content"
        bind:value={content}
        placeholder="Paste your study material to extract flashcards from..."
        rows="5"
        required
      ></textarea>
    </div>

    <div class="row">
      <div class="field">
        <label for="fc-lang">Language</label>
        <select id="fc-lang" bind:value={language}>
          <option value="en">English</option>
          <option value="es">Español</option>
          <option value="fr">Français</option>
          <option value="de">Deutsch</option>
        </select>
      </div>

      <div class="field">
        <label for="fc-count">Number of Cards</label>
        <input id="fc-count" type="number" bind:value={cardCount} min="3" max="40" />
      </div>
    </div>

    <div class="field">
      <span id="difficulty-label" class="field-label">Difficulty</span>
      <div class="difficulty-row" role="radiogroup" aria-labelledby="difficulty-label">
        {#each difficulties as d}
          <button
            type="button"
            class="diff-btn"
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
  {:else}
    <div class="field">
      <span id="cards-label" class="field-label">Flashcards</span>
      <div class="cards-list" role="group" aria-labelledby="cards-label">
        {#each cards as card, i}
          <div class="card-row">
            <span class="card-num">{i + 1}</span>
            <div class="card-fields">
              <input
                type="text"
                placeholder="Question"
                value={card.question}
                oninput={(e) => updateCard(i, 'question', e.target.value)}
              />
              <input
                type="text"
                placeholder="Answer"
                value={card.answer}
                oninput={(e) => updateCard(i, 'answer', e.target.value)}
              />
            </div>
            <button type="button" class="remove-btn" onclick={() => removeCard(i)}>
              ×
            </button>
          </div>
        {/each}
      </div>
      <button type="button" class="add-btn" onclick={addCard}>
        + Add Card
      </button>
    </div>
  {/if}

  <div class="field">
    <label for="fc-pause">Pause Between Q&A</label>
    <div class="pause-control">
      <input
        id="fc-pause"
        type="range"
        bind:value={pauseDuration}
        min="2"
        max="15"
        step="1"
      />
      <span class="pause-value">{pauseDuration}s</span>
    </div>
  </div>

  <div class="actions">
    <label class="preview-toggle">
      <input type="checkbox" bind:checked={previewOnly} />
      <span>Preview only</span>
    </label>
    <button type="submit" class="btn-primary">
      {previewOnly ? 'Preview Cards' : 'Generate Flashcards'}
    </button>
  </div>
</form>

<style>
  .flashcard-form {
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .mode-toggle {
    display: flex;
    gap: 4px;
    background: #f5f3f0;
    padding: 3px;
    border-radius: 8px;
  }

  .mode-btn {
    flex: 1;
    padding: 8px 12px;
    border: none;
    background: transparent;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 500;
    color: #6b6b6b;
    cursor: pointer;
    transition: all 0.2s;
  }

  .mode-btn.active {
    background: #fff;
    color: #1a1a1a;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
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

  textarea, input[type="text"], input[type="number"], select {
    padding: 12px 14px;
    border: 1px solid #e8e5e0;
    border-radius: 10px;
    font-size: 14px;
    font-family: inherit;
    background: #fff;
    color: #2d2d2d;
    transition: border-color 0.2s, box-shadow 0.2s;
  }

  textarea:focus, input[type="text"]:focus, input[type="number"]:focus, select:focus {
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

  .difficulty-row {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
  }

  .diff-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 10px 8px;
    border: 1px solid #e8e5e0;
    border-radius: 10px;
    background: #fff;
    cursor: pointer;
    transition: all 0.2s;
  }

  .diff-btn:hover {
    border-color: #d0cbc3;
  }

  .diff-btn.active {
    border-color: #8b7355;
    background: #f9f5ee;
  }

  .diff-label {
    font-size: 13px;
    font-weight: 600;
    color: #2d2d2d;
  }

  .diff-desc {
    font-size: 10px;
    color: #8a8a8a;
  }

  .cards-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .card-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .card-num {
    flex-shrink: 0;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #f0ece6;
    border-radius: 6px;
    font-size: 11px;
    font-weight: 600;
    color: #8b7355;
  }

  .card-fields {
    flex: 1;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 6px;
  }

  .card-fields input {
    padding: 8px 10px;
    font-size: 13px;
  }

  .remove-btn {
    flex-shrink: 0;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    color: #c0b8ae;
    font-size: 18px;
    cursor: pointer;
    border-radius: 6px;
    transition: all 0.2s;
  }

  .remove-btn:hover {
    background: #f5e6e6;
    color: #c47070;
  }

  .add-btn {
    padding: 10px;
    border: 1px dashed #d8d4cc;
    border-radius: 10px;
    background: transparent;
    color: #8b7355;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .add-btn:hover {
    border-color: #8b7355;
    background: #faf8f5;
  }

  .pause-control {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .pause-control input[type="range"] {
    flex: 1;
    height: 4px;
    border: none;
    background: #e8e5e0;
    border-radius: 2px;
    accent-color: #8b7355;
  }

  .pause-value {
    font-size: 14px;
    font-weight: 600;
    color: #8b7355;
    min-width: 32px;
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

  .btn-primary:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(139, 115, 85, 0.35);
  }
</style>
