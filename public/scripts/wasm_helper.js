const go = new Go();
const loadProgram = (name)=> {
    WebAssembly.instantiateStreaming(
        fetch("asm/" + name + ".wasm"),
        go.importObject
    ).then((result) => {
        go.run(result.instance);
    });
};
