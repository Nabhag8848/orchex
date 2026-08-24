"use strict";

function isPlainObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function invalidInput(parameter, message) {
  return {
    error: {
      type: "validation",
      code: "INVALID_INPUT",
      message,
      retryable: false,
      details: { parameter, runtime: "js" },
    },
  };
}

function validateInput(input) {
  if (!isPlainObject(input)) {
    return invalidInput("input", "input must be an object");
  }
  const keys = Object.keys(input);
  if (keys.length !== 1 || keys[0] !== "data") {
    return invalidInput("input", "input must contain only data");
  }
  if (!isPlainObject(input.data)) {
    return invalidInput("input.data", "input.data must be an object");
  }
  return null;
}

exports.handler = async (event) => {
  const { source, input, timeout_ms: timeoutMs } = event;

  const invalid = validateInput(input);
  if (invalid) {
    return invalid;
  }

  let timer;
  const timeout = new Promise((_, reject) => {
    timer = setTimeout(() => {
      const err = new Error("TIMEOUT");
      err.code = "TIMEOUT";
      reject(err);
    }, timeoutMs);
  });

  try {
    const AsyncFunction = Object.getPrototypeOf(
      async function () {},
    ).constructor;
    const data = await Promise.race([
      Promise.resolve().then(() => new AsyncFunction("input", source)(input)),
      timeout,
    ]);
    return { data };
  } catch (err) {
    if (err && err.code === "TIMEOUT") {
      return {
        error: {
          type: "timeout",
          code: "TIMEOUT",
          message: "function exceeded timeout_ms",
          retryable: false,
          details: { timeout_ms: timeoutMs, runtime: "js" },
        },
      };
    }
    return {
      error: {
        type: "operation",
        code: "RUNTIME_ERROR",
        message: String(err && err.message ? err.message : err),
        retryable: false,
      },
    };
  } finally {
    if (timer) {
      clearTimeout(timer);
    }
  }
};
