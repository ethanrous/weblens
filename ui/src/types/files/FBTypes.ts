export enum DraggingStateT {
    NoDrag, // No dragging is taking place
    InternalDrag, // Dragging is of only internal elements
    InterfaceDrag, // Dragging is of interface elements, resizing panels, etc

    // Dragging is from external source, such as
    // dragging files from your computer over the browser
    ExternalDrag,
}
