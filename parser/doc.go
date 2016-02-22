package parser

/*
 * --------------------------------------------------------------------
 * Configuraion parser:
 * --------------------------------------------------------------------
 *     A Parser method can traverse data definitions according to
 *  the format described in section [2] as BNF. The methods has two
 *  arguments: a buffered stream reader and a callback method.
 *     The callback method is invoked by the parser whenever a para-
 *  meter is encountered. The first argument of the callback method
 *  defines the parameter mode encountered: VAR for parameters (vari-
 *  ables) and LIST for list instances. Parameters always have a name
 *  and (string) value; list are notifying 'start' and 'begin' events
 *  to the callback method. A 'list start' event can have a name for
 *  named lists.
 *     The Parser method is used by worker objects that need to parse
 *  nested and named data structures. The worker is responsible to keep
 *  track of list nesting and so forth; in a sense the methods is
 *  similar to a SAX-based parser callback: it is only informed about
 *  parsing events via the callback.
 *     The callback method can be nicely defined as a 'closure' in a
 *  parsing worker object.
 *
 * --------------------------------------------------------------------
 *  [1] Usage:
 * --------------------------------------------------------------------
 *     A Parser method can traverse data definitions according to
 *  the format described in section [2] as BNF. The methods has two
 *  arguments: a buffered stream reader and a callback method.
 *     The callback method is invoked by the parser whenever a para-
 *  meter is encountered. The first argument of the callback method
 *  defines the parameter mode encountered: VAR for parameters (vari-
 *  ables) and LIST for list instances. Parameters always have a name
 *  and (string) value; list are notifying 'start' and 'begin' events
 *  to the callback method. A 'list start' event can have a name for
 *  named lists.
 *     The Parser method is used by worker objects that need to parse
 *  nested and named data structures. The worker is responsible to keep
 *  track of list nesting and so forth; in a sense the methods is
 *  similar to a SAX-based parser callback: it is only informed about
 *  parsing events via the callback.
 *     The callback method can be nicely defined as a 'closure' in a
 *  parsing worker object.
 *
 * --------------------------------------------------------------------
 *  [2] Data format:
 * --------------------------------------------------------------------
 *
 *      <Data>      ::= <Parameter>
 *                    | <Parameter> [','] <Data>
 *                    ;
 *      <Parameter> ::=  <Name> '=' <Value>
 *                    | [<Name> '='] '{' <List> '}'
 *                    ;
 *      <Name>      ::= [\d][^=]*;
 *      <Value>     ::= [^,}]*;
 *      <List>      ::= <Parameter>
 *                    | <Parameter> ',' <List>
 *                    ;
 *
 *  N.B.: Top-level parameters of form 'name=value' MUST always be
 *  terminated by a comma character (',') except if it is the very
 *  last parameter in a stream!
 * --------------------------------------------------------------------
 *
 * (c) 2010 Bernd Fix   >Y<
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
