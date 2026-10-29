# Spendly Frontend

Next.js frontend for the Spendly expense tracking app.

## Tech Stack

- **Framework**: Next.js 15 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **Testing**: Vitest + React Testing Library
- **State Management**: Zustand (to be added)

## Getting Started

### Prerequisites

- Node.js 20+
- npm 10+

### Setup

1. Install dependencies:
```bash
npm install
```

2. Copy environment variables:
```bash
cp .env.example .env.local
```

3. Start development server:
```bash
npm run dev
```

4. Open [http://localhost:3000](http://localhost:3000)

## Available Scripts

- `npm run dev` — Start development server with hot reload
- `npm run build` — Build for production
- `npm run start` — Start production server
- `npm run lint` — Run ESLint
- `npm test` — Run tests
- `npm run test:watch` — Run tests in watch mode
- `npm run test:coverage` — Run tests with coverage report

## Project Structure

```
src/
├── app/              # Next.js App Router pages
│   ├── layout.tsx    # Root layout (HTML, fonts, global providers)
│   ├── page.tsx      # Home page (/)
│   └── globals.css   # Global styles (Tailwind)
└── components/       # Reusable React components
    ├── Hello.tsx     # Example component
    └── Hello.test.tsx
```

## Testing

Tests use Vitest with React Testing Library. Each component has a corresponding `.test.tsx` file.

```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage
```

## Building for Production

```bash
npm run build
npm run start
```

The build uses Next.js standalone output, optimized for Docker deployment.

## Docker

The Dockerfile uses multi-stage build for optimal image size (~150-200MB):

```bash
# Build image
docker build -t spendly-frontend .

# Run container
docker run -p 3000:3000 spendly-frontend
```

## Learn More

- [Next.js Documentation](https://nextjs.org/docs)
- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/)
- [Vitest](https://vitest.dev/)
- [Tailwind CSS](https://tailwindcss.com/docs)
