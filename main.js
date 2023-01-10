// senk (nano) - bulleted list editor

// create a new node
let node = (text, indent) => ({text, indent})

let editor // top-level editor element
let nodes = [] // document content
let textareas = [] // textarea elements corresponding to each node

// build the editor
let render = () => {
    // clear old content
    editor.replaceChildren()

    // never let the document be empty (otherwise the user would have no place to type in the content)
    if (nodes.length === 0) nodes.push(node("",0))

    let code = "" // generated HTML
    let level = -1 // current indentation level
    // generate the code
    nodes.forEach((n, i) => {
        // indent
        code += "<ul>".repeat(Math.max(0, n.indent - level))
        code += "</ul>".repeat(-Math.min(0, n.indent - level))
        level = n.indent

        code += "<li><textarea data-index='" + i + "' rows='1'>" + nodes[i].text + "</textarea></li>"
    })
    // close remaining tags
    code += "</ul>".repeat(-Math.min(0, 0 - level))

    editor.innerHTML = code
    areas()
}

// moves the cursor to the end if `cursor` is less than zero
let focusArea = (i, cursor) => {
    textareas[i]?.focus()
    if (cursor !== undefined) {
        let c = cursor < 0 ? textareas[i]?.value?.length : cursor
        textareas[i]?.setSelectionRange(c, c)
    }
}

// rebuild the editor and restore focus
let refresh = (focusIndex, focusCursor) => {
    let c = focusCursor ?? textareas[focusIndex]?.selectionEnd
    render()
    if (focusIndex !== undefined) focusArea(focusIndex, c)
}

// attach event handlers to the given textarea
let area = (a) => {
    let index = parseInt(a.dataset.index)
    textareas[index] = a
    a.className = "ta" // mark as initialized

    let update = () => {
        // dynamic height
        a.style.height = "0"
        a.style.height = a.scrollHeight + "px"

        nodes[index].text = a.value

        localStorage.setItem("data", exportNodes())
    }
    update()
    a.oninput = update
    window.addEventListener("resize", update)

    // keyboard input handling
    a.onkeydown = (e) => {
        // break when we handle the event and return when we ignore it
        switch (e.keyCode) {
            case 38: // (ctrl) up
                if (e.shiftKey) return // let the user select in the current field
                if (e.ctrlKey && index !== 0) { // move the node up
                    let temp = nodes[index]
                    nodes[index] = nodes[index - 1]
                    nodes[index - 1] = temp
                    refresh(index - 1, textareas[index].selectionEnd)
                } else {
                    if (index === 0) return // no previous field
                    focusArea(index - 1, textareas[index].selectionEnd) // switch to the previous field
                }
                break
            case 40: // (ctrl) down
                if (e.shiftKey) return // let the user select in the current field
                if (e.ctrlKey && index !== nodes.length - 1) { // move the node down
                    let temp = nodes[index]
                    nodes[index] = nodes[index + 1]
                    nodes[index + 1] = temp
                    refresh(index + 1, textareas[index].selectionEnd)
                } else {
                    if (index === nodes.length - 1) return // no next field
                    focusArea(index + 1, textareas[index].selectionEnd) // switch to the next field
                }
                break
            case 37: // left
                if (e.shiftKey) return // would unexpectedly break the selection
                if (textareas[index].selectionStart !== 0) return
                focusArea(index - 1, -1)
                break
            case 39: // right
                if (e.shiftKey) return // would unexpectedly break the selection
                if (textareas[index].selectionEnd !== textareas[index].value.length) return
                focusArea(index + 1, 0)
                break
            case 13: // (ctrl) enter
                nodes.splice(index + 1, 0, node("", nodes[index].indent)) // insert new node below
                if (!e.ctrlKey) {
                    // move text after the cursor to the new node
                    nodes[index + 1].text = nodes[index].text.slice(textareas[index].selectionEnd)
                    nodes[index].text = nodes[index].text.slice(0, textareas[index].selectionStart)
                }
                refresh(index + 1)
                break
            case 8: // backspace (at the beginning of the line)
                if (textareas[index].selectionEnd === 0) {
                    if (nodes[index].indent === 0) { // merge node with previous
                        if (index === 0) return // no previous node
                        focusArea(index - 1, nodes[index - 1].text.length)
                        nodes[index - 1].text += nodes[index].text
                        nodes.splice(index, 1)
                        refresh(index - 1)
                    } else { // reduce indentation
                        if (e.ctrlKey) {
                            nodes[index].indent = 0
                        } else {
                            nodes[index].indent--
                        }
                        refresh(index)
                    }
                } else {
                    return
                }
                break
            case 32: // space (at the beginning of the line)
            case 9: // (shift) tab
            case 219: // ctrl [
            case 221: // ctrl ]
                if ((e.keyCode === 32 && textareas[index].selectionEnd !== 0) || ((e.keyCode === 219 || e.keyCode === 221) && !e.ctrlKey)) return
                // change indentation level
                nodes[index].indent = Math.max(nodes[index].indent + ((e.keyCode === 9 && e.shiftKey) || e.keyCode === 219 ? -1 : 1), 0)
                refresh(index)
                break
            default:
                return
        }
        e.preventDefault()
    }
}

// initialize all the new textareas
let areas = () => document.querySelectorAll("textarea:not(.ta)").forEach(area)

// text-based format
let exportNodes = () => nodes.map(n => " ".repeat(n.indent) + n.text).join("\n")
let importNodes = t => {
    nodes =  t.split("\n").map(n => {
        let text = n.trimStart()
        return node(text, n.length - text.length)
    })
}

window.onload = () => {
    editor = document.getElementById("editor")
    importNodes(localStorage.getItem("data") ?? "")
    render()
}